package metrics

import "context"

// NOOP Stuff
var _ Meters = (*Noop)(nil)

type Noop struct{}

func (Noop) RecordMessagesReceived(context.Context) {}

func (Noop) RecordMessagesAcked(context.Context) {}

func (Noop) RecordMessagesForwardedToAdapter(context.Context) {}
