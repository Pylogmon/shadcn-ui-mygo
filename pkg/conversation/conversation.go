// Package conversation provides native conversation composition primitives.
// It covers basic attachment actions, bubble/message layout, keyed virtual
// scrolling, markers, expandable traces and controlled multi-step questions.
// Attachment offers opt-in native file-drop callbacks; applications own drop
// validation and upload transport. It does not implement upstream DOM APIs,
// AI SDK adapters, or MessageScroller's turn anchoring/stream pause heuristics.
// Applications own transport, durable state, cancellation and branching.
package conversation
