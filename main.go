package main

import "context"

func main() {
	a := App{}
	a.Initialize(getEnv(context.Background()))
	a.Run(":8000")
}
