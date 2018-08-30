package main

func newLinkKey(r1, r2 RouterID) string {
	return string(int(r1)) + "-" + string(int(r2))
}

func min(n1, n2 int64) int64 {
	if n1 < n2 {
		return n1
	} else {
		return n2
	}
}
