# captive-browser

A more secure, dedicated, Chrome-based captive portal browser that automatically bypasses custom DNS servers.

`captive-browser` detects the DHCP DNS server and runs a SOCKS5 proxy that resolves hostnames through it. Then it starts a Chrome instance in Incognito mode with a separate data directory and waits for it to exit.

[Read more on my blog.](https://blog.filippo.io/captive-browser)

## Installation

You'll need Chrome and Go 1.9 or newer.

```
go get -u github.com/FiloSottile/captive-browser
cp $(go env GOPATH)/src/github.com/FiloSottile/captive-browser/captive-browser-mac-chrome.toml ~/.config/captive-browser.toml
```

Modify `~/.config/captive-browser.toml` if not running on macOS.

To disable the insecure system captive browser [see here](https://github.com/drduh/macOS-Security-and-Privacy-Guide#captive-portal). If that doesn't work, disable SIP (remember to re-enable it), and rename `/System/Library/CoreServices/Captive Network Assistant.app`.

## Usage

Simply run `captive-browser`, log into the captive portal, and then *quit* (⌘Q) the Chrome instance.

If the binary is not found, try `$(go env GOPATH)/bin/captive-browser`.

To configure the browser, open a non-Incognito window (⌘N).
