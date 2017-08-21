package games

// benchServe serve a bench.
// Handle all things related with watching the game. Adding and removing watchers and
// relaying the game information.
func benchServe(joinWatchChCl *JoinWatchChCl, watchingCh <-chan *WatchingChData) {
	watchers := make(map[int]chan<- *WatchingChData)
Loop:
	for {
		select {
		case p := <-joinWatchChCl.Channel:
			_, isFound := watchers[p.ID]
			isDelete := p.SendCh == nil
			if isFound && isDelete {
				delete(watchers, p.ID)
			} else if !isFound && !isDelete {
				watchers[p.ID] = p.SendCh
			}

		case watchingChData, isOpen := <-watchingCh:
			if !isOpen {
				close(joinWatchChCl.Close) //stope join and leave
				if len(watchers) > 0 {
					for _, ch := range watchers {
						close(ch)
					}
				}
				break Loop
			} else {
				if len(watchers) > 0 {
					for _, ch := range watchers {
						ch <- watchingChData
					}
				}
			}
		} //select
	} //for
}
