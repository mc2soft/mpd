package copyobj

func String(s *string) *string {
	if s == nil {
		return nil
	}
	cop := *s

	return &cop
}
func Int64(i *int64) *int64 {
	if i == nil {
		return nil
	}
	cop := *i

	return &cop
}
func UInt64(i *uint64) *uint64 {
	if i == nil {
		return nil
	}
	cop := *i

	return &cop
}
func Bool(b *bool) *bool {
	if b == nil {
		return nil
	}
	cop := *b

	return &cop
}
