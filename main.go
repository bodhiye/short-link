package main

func main() {
	a := App{}
	a.Initialize(getEnv())
	a.Run(":2333")
}
