package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/armon/go-socks5"
)

type UpstreamResolver struct {
	r *net.Resolver
}

func NewUpstreamResolver(upstream string) *UpstreamResolver {
	return &UpstreamResolver{
		r: &net.Resolver{
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				// Redirect all Resolver dials to the upstream.
				return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(upstream, "53"))
			},
		},
	}
}

func (u *UpstreamResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	addrs, err := u.r.LookupIPAddr(ctx, name)
	if err != nil {
		return ctx, nil, err
	}
	if len(addrs) == 0 {
		return ctx, nil, nil
	}
	// Prefer IPv4, like ResolveIPAddr. I can hear Olafur screaming, but the default
	// go-socks5 Resolver uses ResolveIPAddr, and the interface does not allow any better.
	for _, addr := range addrs {
		if addr.IP.To4() != nil {
			return ctx, addr.IP, nil
		}
	}
	return ctx, addrs[0].IP, nil
	// (Why the hell does this *return* a context?)
}

type Config struct {
	SOCKS5Addr string `toml:"socks5-addr"`
	Browser    string
	DHCP       string `toml:"dhcp-dns"`
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: captive-browser config.toml")
	}
	tomlData, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal("Failed to read config:", err)
	}
	var conf Config
	if err := toml.Unmarshal(tomlData, &conf); err != nil {
		log.Fatal("Failed to parse config:", err)
	}

	log.Printf("Obtaining DHCP DNS server...")
	out, err := exec.Command("/bin/sh", "-c", conf.DHCP).Output()
	if err != nil {
		log.Fatal("Failed to execute dhcp-dns:", err)
	}
	match := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}`).Find(out)
	if match == nil {
		log.Fatal("IPs not found in dhcp-dns output.")
	}
	upstream := string(match)

	srv, err := socks5.New(&socks5.Config{
		Resolver: NewUpstreamResolver(upstream),
	})
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		log.Printf("SOCKS5 proxy pointing to DNS %s started at %s...", upstream, conf.SOCKS5Addr)
		log.Fatal(srv.ListenAndServe("tcp", conf.SOCKS5Addr))
	}()

	log.Printf("Starting browser...")
	cmd := exec.Command("/bin/sh", "-c", conf.Browser)
	cmd.Env = append(os.Environ(), "PROXY="+conf.SOCKS5Addr)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Browser exited, shutting down...")
}
