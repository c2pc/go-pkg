package encryptor

import (
	"fmt"
	"os"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/cipher"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "encrypt",
		Short:   "Text Encryptor",
		Example: "encrypt qwerty",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Usage()
				os.Exit(1)
			}

			plaintext := strings.Join(args, " ")

			en, err := cipher.RC4Cipher.Encrypt([]byte(plaintext))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("Plain Text: %s\n", plaintext)
			fmt.Printf("Encrypted: %s\n", string(en))
		},
	}

	return cmd
}
