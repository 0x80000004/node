package interactive

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/mysterium/node/identity"
	tequilapi_client "github.com/mysterium/node/tequilapi/client"
	"io"
	"log"
	"os"
	"strings"
)

// NewCliClient constructs instance of CLI based Mysterium UI
func NewCliClient(historyFile string, tequilaClient *tequilapi_client.Client) *Client {
	return &Client{
		HistoryFile:   historyFile,
		TequilaClient: tequilaClient,
	}
}

// Client describes CLI based Mysterium UI
type Client struct {
	HistoryFile   string
	TequilaClient *tequilapi_client.Client
}

const redColor = "\033[31m%s\033[0m"

// Run starts CLI interface
func (c *Client) Run() error {
	completer := getAutocompleterMenu(c.TequilaClient)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf(redColor, "» "),
		HistoryFile:     c.HistoryFile,
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		return err
	}

	defer rl.Close()

	log.SetOutput(rl.Stderr())

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		c.handleActions(completer, line)
	}
	return nil
}

func (c *Client) handleActions(completer *readline.PrefixCompleter, line string) {
	line = strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(line, "connect"):
		c.connect(completer, line)
		break
	case line == "exit" || line == "quit":
		os.Exit(0)
		break

	case line == "help":
		c.help(completer)
		break

	case line == "status":
		c.status()
		break

	case line == "disconnect":
		c.disconnect()
		break

	case strings.HasPrefix(line, "identities"):
		c.identities(line)
		break

	default:
		if len(line) > 0 {
			c.help(completer)
			break
		}
	}
}

func (c *Client) connect(completer *readline.PrefixCompleter, line string) {
	connectionArgs := strings.TrimSpace(line[7:])
	if len(connectionArgs) == 0 {
		info("Press tab to select identity or create a new one. Connect <your-identity> <node-identity>")
		return
	}

	identities := strings.Fields(connectionArgs)

	if len(identities) != 2 {
		info("Please type in the node identity. Connect <your-identity> <node-identity>")
		return
	}

	clientIdentity, nodeIdentity := identities[0], identities[1]

	if clientIdentity == "new" {
		id, err := c.TequilaClient.NewIdentity()
		if err != nil {
			warn(err)
			return
		}
		clientIdentity = id.Address
		success("New identity created:", clientIdentity)
	}

	status("CONNECTING", "from:", clientIdentity, "to:", nodeIdentity)

	err := c.TequilaClient.Connect(identity.FromAddress(clientIdentity), identity.FromAddress(nodeIdentity))
	if err != nil {
		warn(err)
		return
	}

	success("Connected.")
}

func (c *Client) disconnect() {
	err := c.TequilaClient.Disconnect()
	if err != nil {
		warn(err)
		return
	}

	success("Disconnected.")
}

func (c *Client) status() {
	status, err := c.TequilaClient.Status()
	if err != nil {
		warn(err)
		return
	}

	info("Status:", status.Status)
	info("SID", status.SessionId)
}

func (c *Client) help(completer *readline.PrefixCompleter) {
	info("Mysterium CLI tequilapi commands:")
	fmt.Println(completer.Tree("  "))
}

func (c *Client) identities(line string) {
	action := strings.TrimSpace(line[10:])
	if len(action) == 0 {
		info("identities command:\n    list\n    new")
		return
	}

	if action == "list" {
		ids, err := c.TequilaClient.GetIdentities()
		if err != nil {
			fmt.Println("Error occured:", err)
			return
		}

		for _, id := range ids {
			status("+", id.Address)
		}
		return
	}

	if action == "new" {
		id, err := c.TequilaClient.NewIdentity()
		if err != nil {
			warn(err)
			return
		}
		success("New identity created:", id.Address)
	}
}

func getIdentityOptionList(restClient *tequilapi_client.Client) func(string) []string {
	return func(line string) []string {
		identities := []string{"new"}
		ids, err := restClient.GetIdentities()
		if err != nil {
			warn(err)
			return identities
		}
		for _, id := range ids {
			identities = append(identities, id.Address)
		}

		return identities
	}
}

func getAutocompleterMenu(restClient *tequilapi_client.Client) *readline.PrefixCompleter {
	var completer = readline.NewPrefixCompleter(
		readline.PcItem(
			"connect",
			readline.PcItemDynamic(
				getIdentityOptionList(restClient),
			),
		),
		readline.PcItem(
			"identities",
			readline.PcItem("new"),
			readline.PcItem("list"),
		),
		readline.PcItem("status"),
		readline.PcItem("disconnect"),
		readline.PcItem("help"),
		readline.PcItem("quit"),
	)

	return completer
}
