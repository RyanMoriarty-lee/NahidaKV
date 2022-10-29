package cache

import (
	"container/list"
	"fmt"
)

const (
	STAGE_ONE = iota + 1
	STAGE_TWO
)

type segmentedLRU struct {
	data                     map[uint64]*list.Element
	stageOneCap, stageTowCap int
	stageOne, stageTwo       *list.List
}

func newSLRU(data map[uint64]*list.Element, stageOneCap, stageTowCap int) *segmentedLRU {
	return &segmentedLRU{
		data:        data,
		stageOneCap: stageOneCap,
		stageTowCap: stageTowCap,
		stageOne:    list.New(),
		stageTwo:    list.New(),
	}
}

func (slru *segmentedLRU) Len() int {
	return slru.stageTwo.Len() + slru.stageOne.Len()
}

func (slru *segmentedLRU) add(newitem storeItem) {
	newitem.stage = 1

	if slru.stageOne.Len() < slru.stageOneCap || slru.Len() < slru.stageOneCap+slru.stageTowCap {
		slru.data[newitem.key] = slru.stageOne.PushFront(&newitem)
		return
	}

	e := slru.stageOne.Back()
	item := e.Value.(*storeItem)

	delete(slru.data, item.key)

	*item = newitem
	slru.data[item.key] = e
	slru.stageOne.MoveToFront(e)
}

func (slru *segmentedLRU) get(v *list.Element) {
	item := v.Value.(*storeItem)

	if item.stage == STAGE_TWO {
		slru.stageTwo.MoveToFront(v)
		return
	}

	if slru.stageTwo.Len() < slru.stageTowCap {
		slru.stageOne.Remove(v)
		item.stage = STAGE_TWO
		slru.data[item.key] = slru.stageTwo.PushFront(item)
		return
	}

	back := slru.stageTwo.Back()
	bitem := back.Value.(*storeItem)

	*bitem, *item = *item, *bitem

	bitem.stage = STAGE_TWO
	item.stage = STAGE_ONE

	slru.data[item.key] = v
	slru.data[bitem.key] = back

	slru.stageOne.MoveToFront(v)
	slru.stageTwo.MoveToFront(back)
}

func (slru *segmentedLRU) victim() *storeItem {
	if slru.Len() < slru.stageOneCap+slru.stageTowCap {
		return nil
	}

	v := slru.stageOne.Back()
	return v.Value.(*storeItem)
}

func (slru *segmentedLRU) String() string {
	var s string
	for e := slru.stageTwo.Front(); e != nil; e = e.Next() {
		s += fmt.Sprintf("%v,", e.Value.(*storeItem).value)
	}
	s += fmt.Sprintf(" | ")
	for e := slru.stageOne.Front(); e != nil; e = e.Next() {
		s += fmt.Sprintf("%v,", e.Value.(*storeItem).value)
	}
	return s
}
