# captive-browser

A more secure, dedicated, Chrome-based captive portal browser that automatically bypasses custom DNS servers.

`captive-browser` detects the DHCP DNS server and runs a SOCKS5 proxy that resolves hostnames through it. Then it starts a Chrome instance in Incognito mode with a separate data directory and waits for it to exit.

[Read more on my blog.](https://blog.filippo.io/captive-browser)

## Installation

You'll need Chrome (or Chromium) and Go 1.9 or newer.

```
go get -u github.com/FiloSottile/captive-browser
```

You have to install a config file in `$XDG_CONFIG_HOME/captive-browser.toml` (if set) or `~/.config/captive-browser.toml`. You can probably use one of the stock config below but it might be needed to modify the network interface.
You can also amend the config to use Chromium instead of Chrome (see browser variable).


### macOS

```
cp $(go env GOPATH)/src/github.com/FiloSottile/captive-browser/captive-browser-mac.toml ~/.config/captive-browser.toml
```

To disable the insecure system captive browser [see here](https://github.com/drduh/macOS-Security-and-Privacy-Guide#captive-portal). If that doesn't work, disable SIP (remember to re-enable it), and rename `/System/Library/CoreServices/Captive Network Assistant.app`.

### Linux with NetworkManager (Ubuntu)

```
cp $(go env GOPATH)/src/github.com/FiloSottile/captive-browser/captive-browser-linux-networkmanager.toml ~/.config/captive-browser.toml
```

### Linux with systemd-networkd

```
go get -u github.com/FiloSottile/captive-browser/cmd/systemd-networkd-dns
cp $(go env GOPATH)/src/github.com/FiloSottile/captive-browser/captive-browser-linux-systemd-networkd.toml ~/.config/captive-browser.toml
```

### Linux with dhcpcd

```
cp $(go env GOPATH)/src/github.com/FiloSottile/captive-browser/captive-browser-linux-dhcpd.toml ~/.config/captive-browser.toml
```

## Usage

Simply run `captive-browser`, log into the captive portal, and then *quit* (⌘Q / Ctrl-Q) the Chrome instance.

If the binary is not found, try `$(go env GOPATH)/bin/captive-browser`.

To configure the browser, open a non-Incognito window (⌘N / Ctrl-N).
