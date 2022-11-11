package srvservice

type MyStack struct {
	curPath []string
	size    int
	cap     int
}

func NewStack(basePath string, size, cap int) *MyStack {
	mystack := &MyStack{
		curPath: make([]string, size, cap),
		size:    size,
		cap:     cap,
	}
	mystack.SetBasePath(basePath)
	return mystack
}

func (stack *MyStack) SetBasePath(baseDir string) {
	stack.push(baseDir)
}

func (stack *MyStack) push(path string) {
	if stack.size >= stack.cap {
		return
	}
	stack.curPath = append(stack.curPath, path)
	stack.size++
}

func (stack *MyStack) pop() {
	if stack.size <= 1 {
		return
	}
	stack.size--
}
