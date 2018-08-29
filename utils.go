package main

func newLinkKey(r1, r2 RouterID) string {
	return string(int(r1)) + "-" + string(int(r2))
}

