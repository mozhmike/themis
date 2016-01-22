package gothemis

import (
    "testing"
    "crypto/rand"
    "math/big"
    "bytes"
)

type testCallbacks struct {
	a *Keypair
	b *Keypair
}

func (clb *testCallbacks) GetPublicKeyForId(ss *SecureSession, id []byte) (*PublicKey) {
	switch {
		case 1 == id[0]:
			return clb.a.public
		case 2 == id[0]:
			return clb.b.public	
	}
	
	return nil
}

func (clb *testCallbacks) StateChanged(ss *SecureSession, state int) {
	
}

func genRandData() ([]byte, error) {
	data_length, err := rand.Int(rand.Reader, big.NewInt(2048))
	if nil != err {
		return nil, err
	}
	
	data := make([]byte, int(data_length.Int64()))
	_, err = rand.Read(data)
	if nil != err {
		return nil, err
	}
	
	return data, nil
}

func fin() ([]byte) {
	f := [4]byte{0xDE, 0xAD, 0xC0, 0xDE}
	return f[:]
}

func isFin(b []byte) (bool) {
	return 0 == bytes.Compare(b, fin())
}

func clientService(client *SecureSession, ch chan []byte, finCh chan int, t *testing.T) {
	defer func() {
		if t.Failed() {
			finCh <- 0
		}
	}()
	
	conReq, err := client.ConnectRequest()
	if nil != err {
		t.Error(err)
		return
	}
	
	ch <- conReq
	for {
		buf := <-ch
		
		buf, sendPeer, err := client.Unwrap(buf)
		if nil != err {
			t.Error(err)
			return
		}
		
		if sendPeer {
			ch <- buf
			continue
		}
		
		var finish bool
		if nil == buf {
			buf, _ = genRandData()
		} else {
			buf = fin()
			finish = true
		}
		
		buf, err = client.Wrap(buf)
		if nil != err {
			t.Error(err)
			return
		}
		ch <- buf
		
		if finish {
			break
		}
	}
}

func serverService(server *SecureSession, ch chan []byte, finCh chan int, t *testing.T) {
	defer func() {finCh <- 0}()
	
	for {
		buf := <-ch

		buf, sendPeer, err := server.Unwrap(buf)
		if nil != err {
			t.Error(err)
			return
		}
		
		if !sendPeer {
			if (isFin(buf)) {
				break
			}
			
			buf, err = server.Wrap(buf)
			if nil != err {
				t.Error(err)
				return
			}
		}
		
		ch <- buf
	}
}

func testSession(keytype int, t *testing.T) {
	kpa, err := NewKeypair(keytype)
	if nil != err {
		t.Error(err)
		return
	}
	
	kpb, err := NewKeypair(keytype)
	if nil != err {
		t.Error(err)
		return
	}
	
	clb := &testCallbacks{kpa, kpb}
	
	ida := make([]byte, 1)
	ida[0] = 1
	
	idb := make([]byte, 1)
	idb[0] = 2
	
	client, err := NewSession(ida, kpa.private, clb)
	if nil != err {
		t.Error(err)
		return
	}
	
	server, err := NewSession(idb, kpb.private, clb)
	if nil != err {
		t.Error(err)
		return
	}
	
	ch := make(chan []byte)
	finCh := make(chan int)
	go serverService(server, ch, finCh, t)
	go clientService(client, ch, finCh, t)
	
	<-finCh
}

func TestSession(t *testing.T) {
	testSession(KEYTYPE_EC, t)
}

