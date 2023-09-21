package ad_click

import (
	"context"
	"datagen/gen"
	"datagen/sink"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

type clickEvent struct {
	sink.BaseSinkRecord

	UserId              int64  `json:"user_id"`
	AdId                int64  `json:"ad_id"`
	ClickTimestamp      string `json:"click_timestamp"`
	ImpressionTimestamp string `json:"impression_timestamp"`
}

func (r *clickEvent) Topic() string {
	return "ad_clicks"
}

func (r *clickEvent) Key() string {
	return fmt.Sprint(r.UserId)
}

func (r *clickEvent) ToPostgresSql() string {
	return fmt.Sprintf("INSERT INTO %s (user_id, ad_id, click_timestamp, impression_timestamp) values ('%d', '%d', '%s', '%s')",
		"ad_source", r.UserId, r.AdId, r.ClickTimestamp, r.ImpressionTimestamp)
}

func (r *clickEvent) ToJson() []byte {
	data, _ := json.Marshal(r)
	return data
}

type adClickGen struct {
}

func NewAdClickGen() gen.LoadGenerator {
	return &adClickGen{}
}

func (g *adClickGen) KafkaTopics() []string {
	return []string{"ad_clicks"}
}

func (g *adClickGen) Load(ctx context.Context, outCh chan<- sink.SinkRecord) {
	for {
		now := time.Now()
		record := &clickEvent{
			UserId:              rand.Int63n(100000),
			AdId:                rand.Int63n(10),
			ClickTimestamp:      now.Add(time.Duration(rand.Intn(1000)) * time.Millisecond).Format(gen.RwTimestamptzLayout),
			ImpressionTimestamp: now.Format(gen.RwTimestamptzLayout),
		}
		select {
		case <-ctx.Done():
			return
		case outCh <- record:
		}
	}
}
