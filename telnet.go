package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	//	"strconv"
	"strings"
	"time"
)

type bcmRouteEntry struct {
	index        string
	prefixLen    string
	prefix       string
	nexthopindex string
}

/*
L3_DEFIP_ALPM_IPV6_128.*[2373]: <VALID=0,NEXT_HOP_INDEX=0,LENGTH=0,KEY=0,HIT=0,FLEX_CTR_POOL_NUMBER=0,FLEX_CTR_OFFSET_MODE=0,FLEX_CTR_BASE_COUNTER_IDX=0,EVEN_PARITY=0,ENTRY_ONLY=0,ECMP_PTR=0,ECMP=0,DST_DISCARD=0,DEFAULTROUTE=0,DATA=0,CLASS_ID=0>
*/
var fetchL3DefipAlpmEntry = regexp.MustCompile(`.*\.\*\[(?P<index>[0-9]+)\]\:.*VALID=(?P<valid>[0-9]+)\,.*NEXT_HOP_INDEX=(?P<nexthopindex>[[:word:]]+)\,LENGTH=(?P<masklen>[[:word:]]+)\,KEY=(?P<prefix>[[:word:]]+)\,.*`)

/*
L3_DEFIP.*[1271]: <VRF_ID_MASK1=0,VRF_ID_MASK0=0,VRF_ID_1=0,VRF_ID_0=0,VALID1=0,VALID0=0,SRC_DISCARD1=0,SRC_DISCARD0=0,RPE1=0,RPE0=0,RESERVED_FLEX1=0,RESERVED_FLEX0=0,RESERVED_ECMP_PTR1=0,RESERVED_ECMP_PTR0=0,RESERVED_6=0,RESERVED_5=0,RESERVED_4=0,RESERVED_3=0,RESERVED_2=0,RESERVED_1=0,REPLACE_DATA1=0,REPLACE_DATA0=0,PRI1=0,PRI0=0,NEXT_HOP_INDEX1=0,NEXT_HOP_INDEX0=0,MODE_MASK1=0,MODE_MASK0=0,MODE1=0,MODE0=0,MASK1=0,MASK0=0,KEY1=0,KEY0=0,IP_ADDR_MASK1=0,IP_ADDR_MASK0=0,IP_ADDR1=0,IP_ADDR0=0,HIT1=0,HIT0=0,GLOBAL_ROUTE1=0,GLOBAL_ROUTE0=0,GLOBAL_HIGH1=0,GLOBAL_HIGH0=0,FLEX_CTR_POOL_NUMBER1=0,FLEX_CTR_POOL_NUMBER0=0,FLEX_CTR_OFFSET_MODE1=0,FLEX_CTR_OFFSET_MODE0=0,FLEX_CTR_BASE_COUNTER_IDX1=0,FLEX_CTR_BASE_COUNTER_IDX0=0,EVEN_PARITY=0,ENTRY_VIEW1=0,ENTRY_VIEW0=0,ENTRY_TYPE_MASK1=0,ENTRY_TYPE_MASK0=0,ENTRY_TYPE1=0,ENTRY_TYPE0=0,ECMP_PTR1=0,ECMP_PTR0=0,ECMP1=0,ECMP0=0,D_ID_MASK1=0,D_ID_MASK0=0,D_ID1=0,D_ID0=0,DST_DISCARD1=0,DST_DISCARD0=0,DEFAULT_MISS1=0,DEFAULT_MISS0=0,DEFAULTROUTE1=0,DEFAULTROUTE0=0,CLASS_ID1=0,CLASS_ID0=0,ALG_HIT_IDX1=0,ALG_HIT_IDX0=0,ALG_BKT_PTR1=0,ALG_BKT_PTR0=0>
*/
var fetchL3DefipEntry = regexp.MustCompile(`.*\.\*\[(?P<index>[0-9]+)\]\:.*VALID1=(?P<valid1>[0-9]+)\,VALID0=(?P<valid0>[0-9]+)\,.*NEXT_HOP_INDEX1=(?P<nexthopindex1>[[:word:]]+)\,NEXT_HOP_INDEX0=(?P<nexthopindex0>[[:word:]]+)\,.*IP_ADDR_MASK1=(?P<mask1>[[:word:]]+)\,IP_ADDR_MASK0=(?P<mask0>[[:word:]]+)\,IP_ADDR1=(?P<prefix1>[[:word:]]+)\,IP_ADDR0=(?P<prefix0>[[:word:]]+)\,HIT1=(?P<hit1>[[:word:]]+)\,HIT0=(?P<hit0>[[:word:]]+).*`)

var fetchL3HostEntry = regexp.MustCompile(`.*\.\*\[(?P<index>[0-9]+)\]\:.*VALID=(?P<valid>[0-9]+)\,.*\,NEXT_HOP_INDEX=(?P<nexthopindex>[[:word:]]+)\,.*\,IP_ADDR=(?P<prefix>[[:word:]]+)\,.*`)

func dumpL3HostTable(name string, conn net.Conn) {
	var table_size = 0
	var valid_count = 0

	fmt.Printf("Dumping table: %s\n", name)
	fmt.Println("+++++++++++++++++++++++++++")
	command := "scontrol -f /proc/switch/ASIC/ctrl dump table 0 " + name + "\n"

	_, err := conn.Write([]byte(command))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	reader := bufio.NewReader(conn)
	if reader == nil {
		fmt.Fprintf(os.Stderr, "Create reader failed.")
	}

	for {
		conn.SetReadDeadline(time.Now().Add(time.Duration(3) * time.Second))

		route, err := reader.ReadString('\n')
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				break
			}
			return
		}
		if fetchL3HostEntry.MatchString(route) {
			table_size++
			if fetchL3HostEntry.FindStringSubmatch(route)[2] == "1" ||
				fetchL3HostEntry.FindStringSubmatch(route)[3] == "1" {
				valid_count++
				fmt.Println("\t" + fetchL3HostEntry.FindStringSubmatch(route)[0])
				/*
					fmt.Println(fetchL3HostEntry.FindStringSubmatch(route)[0])
					fmt.Println(fetchL3HostEntry.SubexpNames()[1])
					fmt.Println(fetchL3HostEntry.FindStringSubmatch(route)[1])
					fmt.Println(fetchL3HostEntry.SubexpNames()[2])
					fmt.Println(fetchL3HostEntry.FindStringSubmatch(route)[2])
					fmt.Println(fetchL3HostEntry.SubexpNames()[3])
					fmt.Println(fetchL3HostEntry.FindStringSubmatch(route)[3])
					fmt.Println(fetchL3HostEntry.SubexpNames()[4])
					fmt.Println(fetchL3HostEntry.FindStringSubmatch(route)[4])
				*/

				fmt.Println("+++++++++++++++++++++++++++")
			}
		}
	}

	fmt.Printf("Table (%s) size: %d, entry count: %d\n", name, table_size, valid_count)
}
func dumpDefipTable(name string, conn net.Conn) {
	var table_size = 0
	var valid_count = 0

	fmt.Printf("Dumping table: %s\n", name)
	fmt.Println("+++++++++++++++++++++++++++")
	command := "scontrol -f /proc/switch/ASIC/ctrl dump table 0 " + name + "\n"

	_, err := conn.Write([]byte(command))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	reader := bufio.NewReader(conn)
	if reader == nil {
		fmt.Fprintf(os.Stderr, "Create reader failed.")
	}

	for {
		conn.SetReadDeadline(time.Now().Add(time.Duration(5) * time.Second))

		route, err := reader.ReadString('\n')
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				break
			}
			return
		}
		if fetchL3DefipEntry.MatchString(route) {
			table_size++
			if fetchL3DefipEntry.FindStringSubmatch(route)[2] == "1" ||
				fetchL3DefipEntry.FindStringSubmatch(route)[3] == "1" {
				valid_count++
				fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[0])
				/*
					fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[0])
					fmt.Println(fetchL3DefipEntry.SubexpNames()[1])
					fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[1])
					fmt.Println(fetchL3DefipEntry.SubexpNames()[2])
					fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[2])
					fmt.Println(fetchL3DefipEntry.SubexpNames()[3])
					fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[3])
					fmt.Println(fetchL3DefipEntry.SubexpNames()[4])
					fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[4])
					fmt.Println(fetchL3DefipEntry.SubexpNames()[5])
					fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[5])
					fmt.Println(fetchL3DefipEntry.SubexpNames()[6])
					fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[6])
					fmt.Println(fetchL3DefipEntry.SubexpNames()[7])
					fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[7])
					fmt.Println(fetchL3DefipEntry.SubexpNames()[8])
					fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[8])
					fmt.Println(fetchL3DefipEntry.SubexpNames()[9])
					fmt.Println(fetchL3DefipEntry.FindStringSubmatch(route)[9])
				*/

				fmt.Println("+++++++++++++++++++++++++++")
			}
		}
	}

	fmt.Printf("Table (%s) size: %d, entry count: %d\n", name, table_size, valid_count)
}

func dumpAlpmTable(name string, conn net.Conn) {

	var table_size = 0
	var valid_count = 0

	fmt.Printf("Dumping table: %s\n", name)
	fmt.Println("+++++++++++++++++++++++++++")
	command := "scontrol -f /proc/switch/ASIC/ctrl dump table 0 " + name + "\n"

	_, err := conn.Write([]byte(command))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	reader := bufio.NewReader(conn)
	if reader == nil {
		fmt.Fprintf(os.Stderr, "Create reader failed.")
	}

	for {
		conn.SetReadDeadline(time.Now().Add(time.Duration(5) * time.Second))

		route, err := reader.ReadString('\n')
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				break
			}
			return
		}

		if fetchL3DefipAlpmEntry.MatchString(route) {
			table_size++
			if fetchL3DefipAlpmEntry.FindStringSubmatch(route)[2] == "1" {
				valid_count++
				fmt.Println(fetchL3DefipAlpmEntry.FindStringSubmatch(route)[0])
				/*
					fmt.Println(fetchL3DefipAlpmEntry.FindStringSubmatch(route)[0])
					fmt.Println(fetchL3DefipAlpmEntry.SubexpNames()[1])
					fmt.Println(fetchL3DefipAlpmEntry.FindStringSubmatch(route)[1])
					fmt.Println(fetchL3DefipAlpmEntry.SubexpNames()[2])
					fmt.Println(fetchL3DefipAlpmEntry.FindStringSubmatch(route)[2])
					fmt.Println(fetchL3DefipAlpmEntry.SubexpNames()[3])
					fmt.Println(fetchL3DefipAlpmEntry.FindStringSubmatch(route)[3])
					fmt.Println(fetchL3DefipAlpmEntry.SubexpNames()[4])
					fmt.Println(fetchL3DefipAlpmEntry.FindStringSubmatch(route)[4])
					fmt.Println(fetchL3DefipAlpmEntry.SubexpNames()[5])
					fmt.Println(fetchL3DefipAlpmEntry.FindStringSubmatch(route)[5])
				*/
				fmt.Println("+++++++++++++++++++++++++++")
			}
		}
	}

	fmt.Printf("Table (%s) size: %d, entry count: %d\n", name, table_size, valid_count)
}

func main() {
	//conn, err := net.Dial("tcp", "10.71.20.19:23")
	conn, err := net.Dial("tcp", "10.71.20.187:23")
	if err != nil {
		fmt.Sprint(os.Stderr, "Error: %s", err.Error())
		return
	}

	var buf [4096]byte

	//  for {
	n, err := conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}
	//fmt.Println(string(buf[0:n]))
	//fmt.Println((buf[0:n]))

	buf[1] = 252
	buf[4] = 252
	buf[7] = 252
	buf[10] = 252
	//fmt.Println((buf[0:n]))
	n, err = conn.Write(buf[0:n])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}
	//fmt.Println(string(buf[0:n]))
	//fmt.Println((buf[0:n]))

	buf[1] = 252
	buf[4] = 251
	buf[7] = 252
	buf[10] = 254
	buf[13] = 252
	//fmt.Println((buf[0:n]))
	n, err = conn.Write(buf[0:n])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}
	//fmt.Println(string(buf[0:n]))
	//fmt.Println((buf[0:n]))

	buf[1] = 252
	buf[4] = 252
	//fmt.Println((buf[0:n]))
	n, err = conn.Write(buf[0:n])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}
	//fmt.Println(string(buf[0:n]))
	//fmt.Println((buf[0:n]))

	n, err = conn.Write([]byte("admin\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	//fmt.Println(string(buf[0:n]))

	n, err = conn.Write([]byte("\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}
	//fmt.Println(string(buf[0:n]))

	for {
		n, err = conn.Read(buf[0:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
			return
		}
		//fmt.Println(string(buf[0:n]))
		if strings.HasSuffix(string(buf[0:n]), "> ") {
			break
		}
	}

	n, err = conn.Write([]byte("enable\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}
	//fmt.Println(string(buf[0:n]))

	n, err = conn.Write([]byte("terminal length 0\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}
	//fmt.Println(string(buf[0:n]))

	n, err = conn.Write([]byte("q sh -l\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	reader := bufio.NewReader(conn)
	if reader == nil {
		fmt.Fprintf(os.Stderr, "Create reader failed.")
	}

	for {
		n, err = reader.Read(buf[0:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
			return
		}
		//fmt.Println(string(buf[0:n]))
		if strings.HasSuffix(string(buf[0:n]), "# ") {
			break
		}
	}

	dumpAlpmTable("L3_DEFIP_ALPM_IPV4", conn)
	dumpAlpmTable("L3_DEFIP_ALPM_IPV6_64", conn)
	dumpAlpmTable("L3_DEFIP_ALPM_IPV6_128", conn)
	dumpDefipTable("L3_DEFIP", conn)
	dumpDefipTable("L3_DEFIP_PAIR_128", conn)
	dumpL3HostTable("L3_ENTRY_ONLY", conn)
}
