/*
 * Red-Black tree implementation based on https://github.com/hp685/rbtree
 */

package upf

import (
	//"errors"
	//"sort"
	//"fmt"
	"unsafe"
)

var sentinelPdr *pdr

type queuedPdrsType struct {
	root *pdr
	head *pdr
	qlen int
}

func (queuedPdrs *queuedPdrsType) leftRotate(x *pdr) {
	y := x.right
	x.right = y.left
	if y.left != sentinelPdr {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == sentinelPdr {
		queuedPdrs.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y
}

func (queuedPdrs *queuedPdrsType) rightRotate(y *pdr) {
	x := y.left
	y.left = x.right
	if x.right != sentinelPdr {
		x.right.parent = y
	}
	x.parent = y.parent
	if y.parent == sentinelPdr {
		queuedPdrs.root = x
	} else if y == y.parent.left {
		y.parent.left = x
	} else {
		y.parent.right = x
	}
	x.right = y
	y.parent = x
}

func (queuedPdrs *queuedPdrsType) transplant(u *pdr, v *pdr) {
	if u.parent == sentinelPdr {
		queuedPdrs.root = v
	} else if u == u.parent.left {
		u.parent.left = v
	} else {
		u.parent.right = v
	}
	v.parent = u.parent
}

func (queuedPdrs *queuedPdrsType) insert(node *pdr) {
	//fmt.Printf("→ before insert: %p nextTx: %d\n", node, node.nextTx)
	//debugDump()
	//fmt.Printf("← before insert: %p nextTx: %d\n", node, node.nextTx)
	y := sentinelPdr
	x := queuedPdrs.root
	for {
		if x == sentinelPdr {
			break
		}
		y = x
		if node.lessThan(x) {
			x = x.left
		} else {
			x = x.right
		}
	}
	node.parent = y
	if y == sentinelPdr {
		queuedPdrs.root = node
	} else if node.lessThan(y) {
		y.left = node
	} else {
		y.right = node
	}
	node.left = sentinelPdr
	node.right = sentinelPdr
	node.color = PC_RED
	queuedPdrs.insertFixup(node)
	/* XXX: */
	var head *pdr
	if queuedPdrs.root == sentinelPdr {
		head = nil
	} else {
		head = queuedPdrs.root.treeMinimum()
	}
	if queuedPdrs.head != head {
		queuedPdrs.head = head
	}
	queuedPdrs.qlen++
	//fmt.Printf("→ after insert: %p nextTx: %d\n", node, node.nextTx)
	//debugDump()
	//fmt.Printf("← after insert: %p nextTx: %d\n", node, node.nextTx)
}

func (queuedPdrs *queuedPdrsType) insertFixup(node *pdr) {
	var y *pdr
	for {
		if node.parent.color != PC_RED {
			break
		}
		if node.parent == node.parent.parent.left {
			y = node.parent.parent.right
			if y.color == PC_RED {
				node.parent.color = PC_BLACK
				y.color = PC_BLACK
				node.parent.parent.color = PC_RED
				node = node.parent.parent
			} else {
				if node == node.parent.right {
					node = node.parent
					queuedPdrs.leftRotate(node)
				}
				node.parent.color = PC_BLACK
				node.parent.parent.color = PC_RED
				queuedPdrs.rightRotate(node.parent.parent)
			}
		} else {
			y = node.parent.parent.left
			if y.color == PC_RED {
				node.parent.color = PC_BLACK
				y.color = PC_BLACK
				node.parent.parent.color = PC_RED
				node = node.parent.parent
			} else {
				if node == node.parent.left {
					node = node.parent
					queuedPdrs.rightRotate(node)
				}
				node.parent.color = PC_BLACK
				node.parent.parent.color = PC_RED
				queuedPdrs.leftRotate(node.parent.parent)
			}
		}
	}
	queuedPdrs.root.color = PC_BLACK
}

func (queuedPdrs *queuedPdrsType) remove(node *pdr) {
	//fmt.Printf("→ before remove: %p\n", node)
	//debugDump()
	//fmt.Printf("← before remove: %p\n", node)
	var x *pdr
	y := node
	yOrigColor := y.color
	if node.left == sentinelPdr {
		x = node.right
		queuedPdrs.transplant(node, node.right)
	} else if node.right == sentinelPdr {
		x = node.left
		queuedPdrs.transplant(node, node.left)
	} else {
		y = node.right.treeMinimum()
		yOrigColor = y.color
		x = y.right
		if y.parent == node {
			x.parent = y
		} else {
			queuedPdrs.transplant(y, y.right)
			y.right = node.right
			y.right.parent = y
		}
		queuedPdrs.transplant(node, y)
		y.left = node.left
		y.left.parent = y
		y.color = node.color
	}
	if yOrigColor == PC_BLACK && x != sentinelPdr {
		queuedPdrs.removeFixup(x)
	}
	/* XXX: */
	var head *pdr
	if queuedPdrs.root == sentinelPdr {
		head = nil
	} else {
		head = queuedPdrs.root.treeMinimum()
	}
	if queuedPdrs.head != head {
		queuedPdrs.head = head
	}
	queuedPdrs.qlen--
	//fmt.Printf("→ after remove: %p\n", node)
	//debugDump()
	//fmt.Printf("← after remove: %p\n", node)
}

func (queuedPdrs *queuedPdrsType) removeFixup(node *pdr) {
	var w *pdr
	for {
		if node == queuedPdrs.root || node.color != PC_BLACK {
			break
		}
		if node == node.parent.left {
			w = node.parent.right
			if w.color == PC_RED {
				w.color = PC_BLACK
				node.parent.color = PC_RED
				queuedPdrs.leftRotate(node.parent)
				w = node.parent.right
			}
			if w.left.color == PC_BLACK && w.right.color == PC_BLACK {
				w.color = PC_RED
				node = node.parent
			} else {
				if w.right.color == PC_BLACK {
					w.left.color = PC_BLACK
					w.color = PC_RED
					queuedPdrs.rightRotate(w)
					w = node.parent.right
				}
				w.color = node.parent.color
				node.parent.color = PC_BLACK
				w.right.color = PC_BLACK
				queuedPdrs.leftRotate(node.parent)
				node = queuedPdrs.root
			}
		} else {
			w = node.parent.left
			if w.color == PC_RED {
				w.color = PC_BLACK
				node.parent.color = PC_RED
				queuedPdrs.rightRotate(node.parent)
				w = node.parent.left
			}
			if w.left.color == PC_BLACK && w.right.color == PC_BLACK {
				w.color = PC_RED
				node = node.parent
			} else {
				if w.left.color == PC_BLACK {
					w.right.color = PC_BLACK
					w.color = PC_RED
					queuedPdrs.leftRotate(w)
					w = node.parent.left
				}
				w.color = node.parent.color
				node.parent.color = PC_BLACK
				w.left.color = PC_BLACK
				queuedPdrs.rightRotate(node.parent)
				node = queuedPdrs.root
			}
		}
	}
	node.color = PC_BLACK
}

func (queuedPdrs *queuedPdrsType) exists(pdr *pdr) bool {
	node := queuedPdrs.root
	for {
		if node == sentinelPdr || node == pdr {
			break
		}
		if pdr.lessThan(node) {
			node = node.left
		} else {
			node = node.right
		}
	}
	if node == sentinelPdr {
		return false
	}
	return true
}

func (aPdr *pdr) lessThan(bPdr *pdr) bool {
	aNext := aPdr.nextTx
	bNext := bPdr.nextTx
	if aNext != bNext {
		return int64(aNext-bNext) < 0
	}
	aQold := aPdr.pktq.qold()
	bQold := bPdr.pktq.qold()
	if aQold != bQold {
		return int64(aQold-bQold) < 0
	}
	aUint := uintptr(unsafe.Pointer(aPdr))
	bUint := uintptr(unsafe.Pointer(bPdr))
	return aUint < bUint
}

func (node *pdr) treeMinimum() *pdr {
	for {
		if node.left == sentinelPdr {
			break
		}
		node = node.left
	}
	return node
}

func (node *pdr) treeMaximum() *pdr {
	for {
		if node.right == sentinelPdr {
			break
		}
		node = node.right
	}
	return node
}

func init() {
	sentinelPdr = &pdr{
		color:  PC_BLACK,
		parent: nil,
		left:   nil,
		right:  nil,
	}
	queuedPdrs.root = sentinelPdr
}
