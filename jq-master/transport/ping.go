package transport

import (
	"context"
	"log"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

const (
	RPingSub         = "jq:ping"
	PingDropInterval = 10 * time.Second
)

type PingFailure struct {
	WorkerID string
}

type PingChecker struct {
	conn     *Conn
	workerID string
}

func NewPingChecker(conn *Conn, workerID string) *PingChecker {
	return &PingChecker{conn: conn, workerID: workerID}
}

type pingMessage struct {
	WorkerID string `msgpack:"worker_id"`
}

func (pc *PingChecker) PingCheckLoop(ctx context.Context, onFailure func(PingFailure)) {
	timer := time.NewTicker(PingDropInterval)
	defer timer.Stop()

	sub := pc.conn.r.Subscribe(ctx, RPingSub)
	defer func() {
		err := sub.Close()
		if err != nil {
			log.Printf("error closing subscription: %v", err)
		}
	}()
	subChan := sub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			// timer fired before receiving ping
			// assume that the worker is dead
			pingChan := PingFailure{WorkerID: pc.workerID}
			onFailure(pingChan)
			return
		case msg := <-subChan:
			var ping pingMessage
			if err := msgpack.Unmarshal([]byte(msg.Payload), &ping); err != nil {
				log.Printf("invalid ping message: %v", err)
				continue
			}
			timer.Reset(PingDropInterval)
		}
	}
}
