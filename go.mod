module github.com/TrueBlocks/trueblocks-dalleserver

go 1.23.8

// Use local submodule for active development of trueblocks-dalle
replace github.com/TrueBlocks/trueblocks-dalle/v2 => ./dalle

require (
	github.com/TrueBlocks/trueblocks-core/src/apps/chifra v0.0.0-20250819215522-14adadd17718
	github.com/TrueBlocks/trueblocks-dalle/v2 v2.0.0-00010101000000-000000000000
	golang.org/x/term v0.31.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	git.sr.ht/~sbinet/gg v0.6.0 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bykof/gostradamus v1.1.2 // indirect
	github.com/campoy/embedmd v1.0.0 // indirect
	github.com/crackcomm/go-gitignore v0.0.0-20241020182519-7843d2ba8fdf // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ethereum/go-ethereum v1.15.10 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/ipfs/boxo v0.29.1 // indirect
	github.com/ipfs/go-cid v0.5.0 // indirect
	github.com/ipfs/go-ipfs-api v0.7.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-flow-metrics v0.2.0 // indirect
	github.com/libp2p/go-libp2p v0.41.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.15.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multicodec v0.9.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-multistream v0.6.0 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/panjf2000/ants/v2 v2.11.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/image v0.26.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	lukechampine.com/blake3 v1.4.1 // indirect
)
