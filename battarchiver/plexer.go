package main

type plexer struct {
	chs []string
	ix  int
}

func newPlexer() (px *plexer) {
	px = new(plexer)
	px.chs = make([]string, 0, 10)
	px.ix = -1
	return px
}
func (px *plexer) add(ch string) {
	px.chs = append(px.chs, ch)
	if px.ix == -1 {
		px.ix = 0
	}
}
func (px *plexer) remove(ch string) {
	for i, v := range px.chs {
		if v == ch {
			px.chs = append(px.chs[:i], px.chs[i+1:]...)
			if px.ix > i {
				px.ix = px.ix - 1
			} else if px.ix == i && i == len(px.chs) {
				if px.ix == 0 {
					px.ix = -1
				} else {
					px.ix = 0
				}
			}
			break
		}
	}
}
func (px *plexer) get() (ch string) {
	if px.ix != -1 {
		ch = px.chs[px.ix]
		if len(px.chs) != 1 {
			if len(px.chs)-1 == px.ix {
				px.ix = 0
			} else {
				px.ix = px.ix + 1
			}
		}
	}
	return ch
}
