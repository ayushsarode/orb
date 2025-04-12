package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ayushsarode/orb/internal/index"
	"github.com/ayushsarode/orb/internal/objects"
	"github.com/ayushsarode/orb/internal/refs"
	"github.com/spf13/cobra"
)

func newCommitCommnad() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use: "commit",
		Short: "Record changes to the repository",
		RunE: func (cmd *cobra.Command,args [] string)  error {
			if message == "" {
				return fmt.Errorf("commit message cannot be empty")
			}

			idx, err := index.LoadIndex()
			if err != nil {
				return fmt.Errorf("loading index: %w", err)
			}

			if len(idx.Entries) == 0 {
				return fmt.Errorf("nothing to commit")
			}


			treeContent := buildTreeContent(idx)
			treeHash, err := objects.WriteBlob(objects.TreeType, treeContent)

			if err != nil {
				return fmt.Errorf("writing tree object: %w", err)
			}

			parentHash := ""
			head, err := refs.GetHead()
			if err != nil {
				parentHash  = head
			}

			commitContent := buildCommitContent(treeHash,parentHash, message)
			commitHash, err := objects.WriteObject(objects.CommitType, commitContent)

			if err != nil {
				 return fmt.Errorf("writing commit object: %w", err)
			}

			headContent, err := os.ReadFile("./orb/HEAD")
			if err != nil {
				return fmt.Errorf("reading HEAD: %w", err)
			}

			headRef := string(headContent)

			if len(headRef) > 5 && headRef[:5] == "ref: " {
				branchRef := headRef[5:len(headRef) - 1 ]
				if err := refs.UpdateRef(branchRef, commitHash); err != nil {
					return fmt.Errorf("updating branch reference: %w", err)
				}
			} else {
				if err := refs.UpdateHead(commitHash); err != nil {
					return fmt.Errorf("updating HEAD: %w", err)
				}
			}
			fmt.Printf("[%s] %s\n", commitHash[:7], message)
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message")
	cmd.MarkFlagRequired("message")

	return cmd
}

func buildTreeContent(idx *index.Index) []byte {



	var result []byte

	for _, entry := range idx.GetEntries() {
		line := fmt.Sprintf("100644 blob %s\t%s", entry.ObjectHash, entry.Path)
		result = append(result, []byte(line)...)
		result = append(result, '\n')
	}

	return result
}

func buildCommitContent(treeHash, parentHash, message string) []byte{
	var content string

	content += fmt.Sprintf("tree %s\n", treeHash)

	if parentHash != "" {
		content += fmt.Sprintf("parent %s\n", parentHash)
	}

	timestamp := time.Now().Unix()
	timezone := "+0000"


	author := fmt.Sprintf("Author Name <author@example.com> %d %s", timestamp, timezone)
	committer := fmt.Sprintf("Committer Name <committer@example.com> %d %s", timestamp, timezone)


	content += fmt.Sprintf("author %s\n", author)
	content += fmt.Sprintf("committer %s\n", committer)
	content += fmt.Sprintf("\n%s\n", message)

	return []byte(content)


}