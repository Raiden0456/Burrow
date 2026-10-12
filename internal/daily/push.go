package daily

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/SherClockHolmes/webpush-go"

	"github.com/raiden0456/gonteractive/internal/settings"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

const vapidKey = "vapid"

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

type vapidKeys struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

type Pusher struct {
	q       *db.Queries
	contact string
	keys    vapidKeys
}

func NewPusher(ctx context.Context, q *db.Queries, s *settings.Store, contact string) (*Pusher, error) {
	var keys vapidKeys
	found, err := s.Get(ctx, vapidKey, &keys)
	if err != nil {
		return nil, err
	}
	if !found {
		keys.Private, keys.Public, err = webpush.GenerateVAPIDKeys()
		if err != nil {
			return nil, fmt.Errorf("generate VAPID keys: %w", err)
		}
		if err := s.Put(ctx, vapidKey, keys); err != nil {
			return nil, err
		}
	}
	return &Pusher{q: q, contact: contact, keys: keys}, nil
}

func (p *Pusher) PublicKey() string { return p.keys.Public }

func (p *Pusher) Notify(ctx context.Context, n Notification) (int, error) {
	subs, err := p.q.ListPushSubscriptions(ctx)
	if err != nil {
		return 0, err
	}
	payload, err := json.Marshal(n)
	if err != nil {
		return 0, err
	}

	sent := 0
	for _, sub := range subs {
		resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys:     webpush.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
		}, &webpush.Options{
			Subscriber:      p.contact,
			VAPIDPublicKey:  p.keys.Public,
			VAPIDPrivateKey: p.keys.Private,
			TTL:             12 * 60 * 60,
			Urgency:         webpush.UrgencyNormal,
		})
		if err != nil {
			slog.Warn("push failed", "endpoint", sub.Endpoint, "err", err)
			continue
		}
		resp.Body.Close()
		switch {
		case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
			if err := p.q.DeletePushSubscription(ctx, sub.Endpoint); err != nil {
				return sent, err
			}
		case resp.StatusCode >= 300:
			slog.Warn("push rejected", "endpoint", sub.Endpoint, "status", resp.StatusCode)
		default:
			sent++
		}
	}
	return sent, nil
}
