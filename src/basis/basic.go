package maim

import "fmt"

type Student struct {
	name string
	age  int
}

//在函数声明时，在其名字之前放上一个变量，即是一个方法
func (s Student) pname() {
	fmt.Println(s.name)
}

func main() {
	var elio Student = Student{"elio", 12}
	elio.pname()
}
