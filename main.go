package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	socks5 "github.com/armon/go-socks5"
)

type UpstreamResolver struct {
	r *net.Resolver
}

func NewUpstreamResolver(upstream string) *UpstreamResolver {
	return &UpstreamResolver{
		r: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				// Redirect all Resolver dials to the upstream.
				return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(upstream, "53"))
			},
		},
	}
}

func (u *UpstreamResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	log.Println("Redirected DNS lookup:", name)
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
	NIC        string `toml:"systemd-network-interface"`
}

func main() {
	usr, _ := user.Current()
	configPath := filepath.Join(usr.HomeDir, ".config", "captive-browser.toml")
	tomlData, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalln("Failed to read config:", err)
	}
	var conf Config
	if err := toml.Unmarshal(tomlData, &conf); err != nil {
		log.Fatalln("Failed to parse config:", err)
	}

	var upstream string
	if conf.NIC != "" {
		log.Printf("Obtaining DHCP DNS server from systemd-resolved via DBus...")
		upstream, err = getDHCPDNSForInterfaceFromDBus(conf.NIC)
		if err != nil {
			log.Fatalln("Could not fetch DNS from DBus", err)
		}
		log.Printf("Discovered DNS from DBus: %+v", upstream)
	} else {
		log.Printf("Obtaining DHCP DNS server using dhcp-dns command...")
		out, err := exec.Command("/bin/sh", "-c", conf.DHCP).Output()
		if err != nil {
			log.Fatalln("Failed to execute dhcp-dns:", err)
		}
		match := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}.\d{1,3}`).Find(out)
		if match == nil {
			log.Fatalln("IPs not found in dhcp-dns output.")
		}
		upstream = string(match)
	}

	log.Println("UPSTREAM ------->", upstream)

	srv, err := socks5.New(&socks5.Config{
		Resolver: NewUpstreamResolver(upstream),
	})
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		log.Printf("SOCKS5 proxy pointing to DNS %s started at %s...", upstream, conf.SOCKS5Addr)
		log.Fatalln(srv.ListenAndServe("tcp", conf.SOCKS5Addr))
	}()

	log.Printf("Starting browser...")
	cmd := exec.Command("/bin/sh", "-c", conf.Browser)
	cmd.Env = append(os.Environ(), "PROXY="+conf.SOCKS5Addr)
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
	log.Printf("Browser exited, shutting down...")
}
