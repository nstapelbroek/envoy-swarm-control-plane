package snapshot

type UpdateReason string

// RequestUpdate queues a discovery run without blocking the caller.
// When a run is already queued we skip, as that run will read the latest state anyway.
func RequestUpdate(ch chan UpdateReason, reason UpdateReason) {
	select {
	case ch <- reason:
	default:
	}
}
