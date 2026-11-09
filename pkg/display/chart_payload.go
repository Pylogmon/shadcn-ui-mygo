package display

// JSON-shaped source data is copied without retaining UI contexts or mutating
// the caller. Other application-owned objects are opaque, as in ChartPayloadItem.
func chartCopyPayloadValue(v any) any {
	switch v := v.(type) {
	case map[string]any:
		return chartCopyPayloadRow(v)
	case []any:
		a := make([]any, len(v))
		for i, x := range v {
			a[i] = chartCopyPayloadValue(x)
		}
		return a
	case []string:
		return append([]string(nil), v...)
	case []float64:
		return append([]float64(nil), v...)
	case []int:
		return append([]int(nil), v...)
	case []map[string]any:
		a := make([]map[string]any, len(v))
		for i, x := range v {
			a[i] = chartCopyPayloadRow(x)
		}
		return a
	default:
		return v
	}
}
func chartCopyPayloadRow(v map[string]any) map[string]any {
	if v == nil {
		return nil
	}
	r := make(map[string]any, len(v))
	for k, x := range v {
		r[k] = chartCopyPayloadValue(x)
	}
	return r
}
