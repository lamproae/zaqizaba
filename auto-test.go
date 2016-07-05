package main

import (
	"bufio"
	"fmt"
	//"io"
	"net"
	"os"
	"regexp"
	"strings"
	//"text/template"
	"time"
)

var commandShowIPRoute = "show ip route"
var commandShowIPv6Route = "show ipv6 route"
var commandShowRunningConfig = "show running config"

var commandShowIPOSPF = "show ip ospf"
var commandShowIPOSPFNeighbor = "show ip ospf neighbor"
var commandShowIPOSPFRoute = "show ip ospf route"
var commandShowIPOSPFDatabase = "show ip opsf database"
var commandShowIPOSPFInterface = "show ip ospf interface"
var commandShowIPOSPFVirtualLink = "show ip ospf virtual-links"

var commandShowIPv6OSPF = "show ipv6 ospf"
var commandShowIPv6OSPFNeighbor = "show ipv6 ospf neighbor"
var commandShowIPv6OSPFROute = "show ipv6 ospf route"
var commandShowIPv6OSPFDatabase = "show ipv6 ospf database"
var commandShowIPv6OSPFInterface = "show ipv6 ospf interface"
var commandShowIPv6OSPFVirtualLink = "show ipv6 ospf virtual-links"

var commandShowInterface = "show interface"
var commandShowIPv6Interface = "show ipv6 interface"

type Case struct {
	command  string
	match    *regexp.Regexp
	result   map[string]string
	response string
}

type Tester struct {
	cases   []*Case
	config  *Config
	context *Context
}

var T = &Tester{
	cases: make([]*Case, 0, 100),
	config: &Config{username: "admin",
		password: "",
		address:  "10.71.20.187:23",
	},
}

func (t *Tester) RegisterCase(c *Case) {
	t.cases = append(t.cases, c)
}

func (t *Tester) Run() {
	for _, c := range t.cases {
		fmt.Println(c)
		t.RunCase(c)
	}
}

func (t *Tester) RunCase(c *Case) {
	fmt.Println(c.command)
	fmt.Println(t)
	_, err := t.context.writer.WriteString(c.command + "\n")
	if err != nil {
		fmt.Printf("Send command %s error: %s", c.command, err.Error())
		return
	}

	t.context.writer.Flush()
	for {
		t.context.conn.SetReadDeadline(time.Now().Add(time.Duration(5) * time.Second))

		line, err := t.context.reader.ReadString('\n')
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				break
			}
			return
		}
		c.response += line
	}

	fmt.Println(c.response)
}

type OSPF struct {
	version string
}

type OSPFInterface struct {
	name      string
	neighbors []*OSPFNeighbor
	flags     string
}

type OSPFNeighbor struct {
	address string
	state   string
}

type OSPFAreaType struct {
	neighbors  []*OSPFNeighbor
	interfaces []*OSPFInterface
}

type Config struct {
	username string
	password string
	address  string // ip:port
}

type Context struct {
	conn   net.Conn
	cmode  string
	reader *bufio.Reader
	writer *bufio.Writer
}

func login(config *Config) error {
	var context = &Context{}
	fmt.Println(config.address)
	conn, err := net.DialTimeout("tcp", config.address, time.Duration(3)*time.Second)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return err
	}

	var buf [4096]byte

	//  for {
	n, err := conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return err
	}
	fmt.Println(string(buf[0:n]))
	fmt.Println((buf[0:n]))

	buf[1] = 252
	buf[4] = 252
	buf[7] = 252
	buf[10] = 252
	//fmt.Println((buf[0:n]))
	n, err = conn.Write(buf[0:n])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return err
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return err
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
		return err
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return err
	}
	fmt.Println(string(buf[0:n]))
	fmt.Println((buf[0:n]))

	buf[1] = 252
	buf[4] = 252
	//fmt.Println((buf[0:n]))
	n, err = conn.Write(buf[0:n])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return err
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return err
	}
	//fmt.Println(string(buf[0:n]))
	//fmt.Println((buf[0:n]))

	n, err = conn.Write([]byte(config.username + "\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return err
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return err
	}

	fmt.Println(string(buf[0:n]))

	n, err = conn.Write([]byte(config.password + "\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return err
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return err
	}
	//fmt.Println(string(buf[0:n]))

	for {
		n, err = conn.Read(buf[0:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
			return err
		}
		//fmt.Println(string(buf[0:n]))
		if strings.HasSuffix(string(buf[0:n]), "> ") {
			break
		}
	}

	n, err = conn.Write([]byte("enable\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return err
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return err
	}
	//fmt.Println(string(buf[0:n]))

	n, err = conn.Write([]byte("terminal length 0\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return err
	}

	n, err = conn.Read(buf[0:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return err
	}

	fmt.Println(string(buf[0:n]))
	reader := bufio.NewReader(conn)
	if reader == nil {
		fmt.Printf("Create reader failed.")
	}

	writer := bufio.NewWriter(conn)
	if writer == nil {
		fmt.Printf("Create reader failed.")
	}

	context.conn = conn
	context.reader = reader
	context.writer = writer
	T.context = context
	fmt.Println(T)
	return nil
}

func getAllOSPFConfiguration(conn net.Conn) {
	// dump all process
	// dump all interface
	// dump all neighbors
	// dump all route
	// dump all ospf LSA
}

func main() {
	err := login(T.config)
	if err != nil {
		fmt.Printf("Login error with message: %s\n", err.Error())
		return
	}

	T.RegisterCase(&Case{command: commandShowIPOSPF})
	T.RegisterCase(&Case{command: commandShowIPRoute})
	T.RegisterCase(&Case{command: commandShowInterface})
	T.RegisterCase(&Case{command: commandShowIPv6Interface})
	T.Run()
}

func checkNetworkReachability(address string) error {

	return nil
}

type Test struct {
	S string
}

func init() {

}
