package main

import "crontab/worker"

func main() {
	worker.RunWorker()
	//cmd := exec.CommandContext(context.TODO(), "/bin/bash", "-c", "echo 2")
	//
	//// 执行并捕获输出
	//output, err := cmd.CombinedOutput()
	//fmt.Println(string(output))
	//fmt.Println(err)
}
