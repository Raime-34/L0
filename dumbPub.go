package main

import (
	"bufio"
	"log"
	"os"
)

func main() {

	clusterID := "dumbass"
	sc, conErr := connect2Stan(clusterID)
	if conErr != nil {
		log.Fatalln(conErr)
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		line, _ := reader.ReadString('\n')
		sc.Publish(subject, []byte(line))
	}

}
