package rocksdbkvs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/coreos/etcd/raft/raftpb"
	"github.com/getumen/doctrine/phalanx"
	_ "github.com/getumen/doctrine/phalanx/stablestore/rocksdb"
	"golang.org/x/xerrors"
)

const regionName = "region-1"

type cluster struct {
	peers        []string
	commitC      []<-chan []byte
	errorC       []<-chan error
	proposeC     []chan []byte
	confChangeC  []chan raftpb.ConfChange
	stableStores []phalanx.StableStore
}

// newCluster creates a cluster of n nodes
func newCluster(n int) (*cluster, error) {

	peers := make([]string, n)
	for i := range peers {
		peers[i] = fmt.Sprintf("http://127.0.0.1:%d", 10010+i)
	}

	clus := &cluster{
		peers:        peers,
		commitC:      make([]<-chan []byte, len(peers)),
		errorC:       make([]<-chan error, len(peers)),
		proposeC:     make([]chan []byte, len(peers)),
		confChangeC:  make([]chan raftpb.ConfChange, len(peers)),
		stableStores: make([]phalanx.StableStore, len(peers)),
	}

	var err error

	if err = os.Mkdir("data", 0755); err != nil && !os.IsExist(err) {
		log.Fatalf("fail to create data dir: %+v", err)
	}

	for i := range clus.peers {
		os.RemoveAll(fmt.Sprintf("data/wal-%d", i+1))
		os.RemoveAll(fmt.Sprintf("data/snap-%d", i+1))
		os.RemoveAll(fmt.Sprintf("data/stableStore-%d", i+1))
		clus.proposeC[i] = make(chan []byte, 1)
		clus.confChangeC[i] = make(chan raftpb.ConfChange, 1)
		clus.stableStores[i], err = phalanx.NewStableStore(
			"rocksdb",
			fmt.Sprintf("data/stableStore-%d", i+1),
		)
		if err != nil {
			return nil, err
		}
		clus.stableStores[i].CreateRegion(regionName)
		getSnapshot := func() ([]byte, error) { return clus.stableStores[i].CreateCheckpoint(regionName) }

		clus.commitC[i], clus.errorC[i], _ = phalanx.NewNode(
			i+1,
			clus.peers,
			false,
			getSnapshot,
			clus.proposeC[i],
			clus.confChangeC[i],
			fmt.Sprintf("data/wal-%d", i+1),
			fmt.Sprintf("data/snap-%d", i+1),
		)
	}

	return clus, nil
}

// sinkReplay reads all commits in each node's local log
func (clus *cluster) sinkReplay() {
	for i := range clus.peers {
		for s := range clus.commitC[i] {
			if s == nil {
				break
			}
		}
	}
}

// Close closes all cluster nodes and returns an error if any failed.
func (clus *cluster) Close() (err error) {
	for i := range clus.peers {
		close(clus.proposeC[i])
		for range clus.commitC[i] {
			// drain pending commits
		}
		// wait for channel to close
		if erri := <-clus.errorC[i]; erri != nil {
			err = erri
			return err
		}
		// close stable store
		err = clus.stableStores[i].Close()
		if err != nil {
			return xerrors.Errorf("fail to close stable store %d: %w", i, err)
		}
		// clean intermediates
		os.RemoveAll(fmt.Sprintf("data/wal-%d", i+1))
		os.RemoveAll(fmt.Sprintf("data/snap-%d", i+1))
		os.RemoveAll(fmt.Sprintf("data/stableStore-%d", i+1))

	}
	return err
}

func (clus *cluster) closeNoErrors(t *testing.T) {
	if err := clus.Close(); err != nil {
		t.Fatalf("error: %+v", err)
	}
}

// TestProposeOnCommit starts three nodes and feeds commits back into the proposal
// channel. The intent is to ensure blocking on a proposal won't block raft progress.
func TestProposeOnCommit(t *testing.T) {
	clus, err := newCluster(3)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	defer clus.closeNoErrors(t)

	clus.sinkReplay()

	donec := make(chan struct{})
	for i := range clus.peers {
		// feedback for "n" committed entries, then update donec
		go func(pC chan<- []byte, cC <-chan []byte, eC <-chan error) {
			for n := 0; n < 100; n++ {
				s, ok := <-cC
				if !ok {
					pC = nil
				}
				select {
				case pC <- s:
					continue
				case err := <-eC:
					log.Fatalf("eC message (%+v)", err)
				}
			}
			donec <- struct{}{}
			for range cC {
				// acknowledge the commits from other nodes so
				// raft continues to make progress
			}
		}(clus.proposeC[i], clus.commitC[i], clus.errorC[i])

		// one message feedback per node
		go func(i int) { clus.proposeC[i] <- []byte("foo") }(i)
	}

	for range clus.peers {
		<-donec
	}
}

// TestCloseProposerBeforeReplay tests closing the producer before raft starts.
func TestCloseProposerBeforeReplay(t *testing.T) {
	clus, err := newCluster(1)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	// close before replay so raft never starts
	defer clus.closeNoErrors(t)
}

// TestCloseProposerInflight tests closing the producer while
// committed messages are being published to the client.
func TestCloseProposerInflight(t *testing.T) {
	clus, err := newCluster(1)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	defer clus.closeNoErrors(t)

	clus.sinkReplay()

	// some inflight ops
	go func() {
		clus.proposeC[0] <- []byte("foo")
		clus.proposeC[0] <- []byte("bar")
	}()

	// wait for one message
	if c, ok := <-clus.commitC[0]; !bytes.Equal(c, []byte("foo")) || !ok {
		t.Fatalf("Commit failed")
	}
}

func TestPutAndGetKeyValue(t *testing.T) {

	if err := os.Mkdir("data", 0755); err != nil && !os.IsExist(err) {
		t.Fatalf("fail to create data dir: %+v", err)
	}

	os.RemoveAll(fmt.Sprintf("data/wal-%d", 1))
	os.RemoveAll(fmt.Sprintf("data/snap-%d", 1))
	os.RemoveAll(fmt.Sprintf("data/stableStore-%d", 1))

	t.Cleanup(func() {
		os.RemoveAll(fmt.Sprintf("data/wal-%d", 1))
		os.RemoveAll(fmt.Sprintf("data/snap-%d", 1))
		os.RemoveAll(fmt.Sprintf("data/stableStore-%d", 1))
	})

	clusters := []string{"http://127.0.0.1:9022"}

	proposeC := make(chan []byte)
	defer close(proposeC)

	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	stableStore, err := phalanx.NewStableStore(
		"rocksdb",
		fmt.Sprintf("data/stableStore-%d", 1),
	)
	if err != nil {
		t.Fatalf("fail to create stable store: %+v", err)
	}
	stableStore.CreateRegion(regionName)
	getSnapshot := func() ([]byte, error) { return stableStore.CreateCheckpoint(regionName) }
	commitC, errorC, snapshotterReady := phalanx.NewNode(
		1,
		clusters,
		false,
		getSnapshot,
		proposeC,
		confChangeC,
		fmt.Sprintf("data/wal-%d", 1),
		fmt.Sprintf("data/snap-%d", 1),
	)

	kvs := phalanx.NewDB(
		regionName,
		<-snapshotterReady,
		proposeC,
		commitC,
		errorC,
		stableStore,
		&commandHandler{},
	)

	srv := httptest.NewServer(&httpKVAPI{
		regionName:  regionName,
		store:       kvs,
		confChangeC: confChangeC,
	})
	defer srv.Close()

	// wait server started
	<-time.After(time.Second * 3)

	wantKey, wantValue := []byte("test-key"), []byte("test-value")
	url := fmt.Sprintf("%s/%s", srv.URL, wantKey)
	body := bytes.NewBuffer(wantValue)
	cli := srv.Client()

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/html; charset=utf-8")
	_, err = cli.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// wait for a moment for processing message, otherwise get would be failed.
	<-time.After(time.Second)

	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if !bytes.Equal(wantValue, data) {
		t.Fatalf("expect %s, got %s", wantValue, data)
	}
}
