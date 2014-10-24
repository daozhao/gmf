package main

import (
	"fmt"
	. "github.com/3d0c/gmf"
	"log"
	"os"
	"strconv"
	"time"
	"sync/atomic"
	"os/signal"
	"syscall"
)

var runingCount int32
var cancelRun bool = false

func runClient(url string,no int,exitChan chan int) {

	defer func() {
		exitChan <- no
	}()

	inputCtx,err := NewInputCtx(url)
	if err != nil {
		fmt.Println("Cannot open[",no,"]Url:",url," errorMsg:",err)
		return
	}
	defer inputCtx.CloseInputAndRelease()

//	inputCtx.Dump()

	fmt.Println("===================================")
	var ptsInfo[10] int
	i := 1
	for packet := range inputCtx.GetNewPackets() {

		i++
		if 0 != ptsInfo[packet.StreamIndex()] {

			missPacketCount := 0
			if packet.Duration() != 0 {
				missPacketCount = (packet.Pts()-ptsInfo[packet.StreamIndex()])/packet.Duration()
			}
			if ptsInfo[packet.StreamIndex()] + 2 <= packet.Pts() &&  i > 200 { //这里需要大于200的包是因为200个包之前一般都缺包，不清楚为什么。
				fmt.Print("[", no ,"]Stream[",packet.StreamIndex(),"] i=",i," miss packet. need duration:",
					packet.Pts() - ptsInfo[packet.StreamIndex()]," miss packetCount:",missPacketCount )
//				packet.DumpAtLine()
			}

		}
		ptsInfo[packet.StreamIndex()] = packet.Pts() + packet.Duration()

		Release(packet)

		if cancelRun {
			break
		}
	}
}

func main() {
	var srcFileName string
	var count int
	var exitChan chan int
	exitChan = make(chan int);

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	if len(os.Args) < 3 {
		fmt.Println("usage:",os.Args[0] ," RtmpUrl count")
		fmt.Println("API example program to remux a media file with libavformat and libavcodec.")
		fmt.Println("The output format is guessed according to the file extension.")

		os.Exit(0)
	} else {
		srcFileName = os.Args[1]
		i, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		count = i
	}

	fmt.Println("Run test,url:",srcFileName," count:",count)
//	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		for s := range sigc {
			fmt.Println("Recieve sigc:",s)
			cancelRun = true
			//os.Exit(0)
		}
	}()


	runingCount = 0
	go func() {
		for i:=0 ; i < count ; i++ {
			//		time.Sleep(time.Millisecond *  time.Duration(r.Int63n(50)))
			time.Sleep(time.Second )
			atomic.AddInt32(&runingCount,1)
			go runClient(srcFileName,i,exitChan)
			if cancelRun {
				break
			}
		}
	}()


	for {
		select {
		case  no := <- exitChan :
			fmt.Println("Thread(",no,") is exit. Now is ",runingCount,"'s thread running")
			atomic.AddInt32(&runingCount,-1)
			if runingCount <= 0 {
				fmt.Println("Program will be exit. count:",runingCount)
				os.Exit(0)
			}
		case <-time.After(time.Second * 2) :
			fmt.Println("Program is running. count:",runingCount," at time",time.Now())
		}

	}



}
