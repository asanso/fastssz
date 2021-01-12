package tests

import (
	"errors"

	ssz "github.com/ferranbt/fastssz"
)

type Hash []byte

type Metadata struct {
	Version    uint8
	CodeHash   Hash `ssz-size:"32"`
	CodeLength uint16
}

type Chunk struct {
	FIO  uint8
	Code []byte `ssz-size:"32"` // Last chunk is right-padded with zeros
}

type CodeTrieSmall struct {
	Metadata *Metadata
	Chunks   []*Chunk `ssz-max:"4"`
}

type CodeTrieBig struct {
	Metadata *Metadata
	Chunks   []*Chunk `ssz-max:"1024"`
}

func (md *Metadata) GetTree() (*ssz.Node, error) {
	leaves := md.getLeaves()
	return ssz.TreeFromNodes(leaves)
}

func (md *Metadata) getLeaves() []*ssz.Node {
	leaves := make([]*ssz.Node, 4)
	leaves[0] = ssz.LeafFromUint8(md.Version)
	leaves[1] = ssz.LeafFromBytes(md.CodeHash)
	leaves[2] = ssz.LeafFromUint16(md.CodeLength)
	leaves[3] = ssz.EmptyLeaf()
	return leaves
}

func (t *CodeTrieSmall) GetTree() (*ssz.Node, error) {
	leaves := make([]*ssz.Node, 2)
	// Metadata tree
	mdTree, err := t.Metadata.GetTree()
	if err != nil {
		return nil, err
	}
	leaves[0] = mdTree
	chunkMixinTree, err := t.getChunkListTree()
	if err != nil {
		return nil, err
	}
	leaves[1] = chunkMixinTree
	// Tree with metadata and chunks subtrees
	return ssz.TreeFromNodes(leaves)
}

func (t *CodeTrieSmall) getChunkListTree() (*ssz.Node, error) {
	return getChunkListTree(4, t.Chunks)
}

func getChunkListTree(size int, chunks []*Chunk) (*ssz.Node, error) {
	// Construct a tree  for each chunk
	if len(chunks) > size {
		return nil, errors.New("Number of chunks exceeds capacity")
	}

	chunkTrees := make([]*ssz.Node, len(chunks))
	for i, c := range chunks {
		t, err := c.GetTree()
		if err != nil {
			return nil, err
		}
		chunkTrees[i] = t
	}

	return ssz.TreeFromNodesWithMixin(chunkTrees, size)
}

func (t *CodeTrieBig) GetTree() (*ssz.Node, error) {
	leaves := make([]*ssz.Node, 2)
	// Metadata tree
	mdTree, err := t.Metadata.GetTree()
	if err != nil {
		return nil, err
	}
	leaves[0] = mdTree
	chunkMixinTree, err := t.getChunkListTree()
	if err != nil {
		return nil, err
	}
	leaves[1] = chunkMixinTree
	// Tree with metadata and chunks subtrees
	return ssz.TreeFromNodes(leaves)
}

func (t *CodeTrieBig) getChunkListTree() (*ssz.Node, error) {
	return getChunkListTree(1024, t.Chunks)
}

func (c *Chunk) GetTree() (*ssz.Node, error) {
	leaves := c.getLeaves()
	return ssz.TreeFromNodes(leaves)
}

func (c *Chunk) getLeaves() []*ssz.Node {
	leaves := make([]*ssz.Node, 2)
	leaves[0] = ssz.LeafFromUint8(c.FIO)
	leaves[1] = ssz.LeafFromBytes(c.Code)
	return leaves
}
