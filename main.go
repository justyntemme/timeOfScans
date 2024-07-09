package main

import (
	"fmt"
	"sync"

  "github.com/justyntemme/timeOfScans/params"
  "github.com/justyntemme/timeOfScans/net"
)



func main() {
	paramKeys := []string{"PC_IDENTITY", "PC_SECRET", "TL_URL"}
  paramValues, err := params.GetEnvVars(paramKeys)	
  if err != nil {
		fmt.Println(err)
		return
	}

		cwpToken, err := net.GenerateCwpToken(paramValues["PC_IDENTITY"], paramValues["PC_SECRET"], paramValues["TL_URL"])

	if err != nil {
		fmt.Println(err)
		return
	}

	var wg sync.WaitGroup
	result := make(chan map[string]int)

	wg.Add(1)
	go net.GetAllScansWithTimeCounts(cwpToken, 5, result, &wg, paramValues["TL_URL"])

	go func() {
		wg.Wait()
		close(result)
	}()

	timeCounts := make(map[string]int)
	for counts := range result {
		for key, count := range counts {
			timeCounts[key] += count
		}
	}

	fmt.Println(timeCounts)
}
