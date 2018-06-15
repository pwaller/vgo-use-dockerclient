# Non-optimal vgo experience.

This is a quick note about an experience I have had with `vgo` now on a couple
of projects, whereby it takes a very long time to run, much longer than it took
to run when we were using git submodules or glide vendoring.

To demonstrate the issue I have made a minimal go project which imports
`docker/docker/client`, which is in this repository.

My best guess as to why this is happening is because vgo is visiting the
transitive dependencies of the module root `docker/docker` and not those of
`docker/docker/client`.

To be fair to vgo, I think this behaviour can be explained in part because
`docker/docker` isn't yet a vgo package, so the pain below may only be a
teething pain which won't exist in the end. But I'm unsure on that point.

Questions:

1. Why does `vgo` take 8 minutes to run the first time? I would expect it to
take as long as would be needed to download the `docker/docker` source code zip,
which is 10 megabytes.

2. Why is `vgo` producing so many `exit status 1` errors? Is this expected?

3. Should `vgo` be looking at all of the transitive dependencies of
`docker/docker`? My belief is no, just those of `docker/docker/client`.

---

`vgo` takes a very long time to run for a simple program which imports a docker
client package. The `docker/client` package itself does not have many
dependencies, though obviously `docker/docker` does have many. My expectation is
that vgo could be fast in this case, since it only needs to visit the
dependencies of the client.

In the example below it took 8 minutes to generate a `go.mod`. This was running
on a machine with a very good network connection. From a co-working space it
took hours.

Furthermore, vgo picks up docker `v1.13.1` and not the latest, `v17.05.0-ce`,
presumably because the `-ce` is not recognized as a valid suffix for semantic
versioning.

Once it succeeds, the `go.mod` file is modest, and subsequent builds seem reasonably fast:

```go.mod
module github.com/pwaller/use-dockerclient

require (
	github.com/docker/docker v1.13.1
	github.com/docker/go-units v0.3.3
	github.com/opencontainers/runc v1.0.0-rc5
)
```

However, updating the docker version again causes `vgo install` to take a very
long time.

---

This package contains a very simple program with one dependency.

```go
package main

import (
	"log"

	"github.com/docker/docker/client"
)

func main() {
	c, err := client.NewEnvClient()
	log.Println(c, err)
}
```

If we look at `github.com/docker/docker/client`, it doesn't have many transitive
dependencies, and in fact all of them are available within the repository under
the vendor directory:

```
$ go list -f '{{join .Deps "\n"}}' github.com/docker/docker/client | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}'
github.com/docker/docker/api/types
github.com/docker/docker/api/types/blkiodev
github.com/docker/docker/api/types/container
github.com/docker/docker/api/types/events
github.com/docker/docker/api/types/filters
github.com/docker/docker/api/types/mount
github.com/docker/docker/api/types/network
github.com/docker/docker/api/types/reference
github.com/docker/docker/api/types/registry
github.com/docker/docker/api/types/strslice
github.com/docker/docker/api/types/swarm
github.com/docker/docker/api/types/time
github.com/docker/docker/api/types/versions
github.com/docker/docker/api/types/volume
github.com/docker/docker/pkg/tlsconfig
github.com/docker/docker/vendor/github.com/Sirupsen/logrus
github.com/docker/docker/vendor/github.com/docker/distribution/digest
github.com/docker/docker/vendor/github.com/docker/distribution/reference
github.com/docker/docker/vendor/github.com/docker/go-connections/nat
github.com/docker/docker/vendor/github.com/docker/go-connections/sockets
github.com/docker/docker/vendor/github.com/docker/go-connections/tlsconfig
github.com/docker/docker/vendor/github.com/docker/go-units
github.com/docker/docker/vendor/github.com/opencontainers/runc/libcontainer/user
github.com/docker/docker/vendor/github.com/pkg/errors
github.com/docker/docker/vendor/golang.org/x/net/context
github.com/docker/docker/vendor/golang.org/x/net/context/ctxhttp
github.com/docker/docker/vendor/golang.org/x/net/proxy
```

However, if we run `vgo install -v` with no `go.mod` file, it take 8 minutes and produces a lot of error output. Additionally, it visits a lot of packages which are not used by `github.com/docker/docker/client`. This is surprising to me, as I would expect it to only consider dependencies of `github.com/docker/docker/client`, not of `github.com/docker/docker`.

```
$ time vgo install .
vgo: creating new go.mod: module github.com/pwaller/use-dockerclient
vgo: resolving import "github.com/docker/docker/client"
vgo: stat github.com/Azure/go-ansiterm@388960b655244e76e24c75f48631564eaefade62: git fetch --depth=1 origin 388960b655244e76e24c75f48631564eaefade62 in /home/pwaller/.local/src/v/cache/vcswork/da5145fda272732cd74527dacbe4967cc6a648dad514ccb83c288a28ea4c0671: exit status 1
vgo: stat github.com/davecgh/go-spew@6d212800a42e8ab5c146b8ace3490ee17e5225f9: git fetch --depth=1 origin 6d212800a42e8ab5c146b8ace3490ee17e5225f9 in /home/pwaller/.local/src/v/cache/vcswork/b9a4b9bbdb4a59723f2348415ad7ffda91568455a1cfd92e97976132bdfbaf57: exit status 1
vgo: stat github.com/docker/libtrust@9cbd2a1374f46905c68a4eb3694a130610adc62a: git fetch --depth=1 origin 9cbd2a1374f46905c68a4eb3694a130610adc62a in /home/pwaller/.local/src/v/cache/vcswork/da07fa0f90047c1906572bf6085c832cc0d3e6a5b251b2c1daacbbe31dba899f: exit status 1
vgo: stat github.com/go-check/check@4ed411733c5785b40214c70bce814c3a3a689609: git fetch --depth=1 origin 4ed411733c5785b40214c70bce814c3a3a689609 in /home/pwaller/.local/src/v/cache/vcswork/8a10bcc540d5cc49b62e7151b7bf7502646412570399a1a93fbc8c13a5923bf0: exit status 1
vgo: stat golang.org/x/net@2beffdc2e92c8a3027590f898fe88f69af48a3f8: git fetch --depth=1 origin 2beffdc2e92c8a3027590f898fe88f69af48a3f8 in /home/pwaller/.local/src/v/cache/vcswork/4a22365141bc4eea5d5ac4a1395e653f2669485db75ef119e7bbec8e19b12a21: exit status 128:
	fatal: expected shallow/unshallow, got ERR internal server error
	fatal: The remote end hung up unexpectedly
vgo: stat github.com/docker/go-units@8a7beacffa3009a9ac66bad506b18ffdd110cf97: git fetch --depth=1 origin 8a7beacffa3009a9ac66bad506b18ffdd110cf97 in /home/pwaller/.local/src/v/cache/vcswork/9f2d1a527162210923891d0d8f65d59f937c6335cd9cad946f750ba191817910: exit status 1
vgo: stat github.com/RackSec/srslog@456df3a81436d29ba874f3590eeeee25d666f8a5: git fetch --depth=1 origin 456df3a81436d29ba874f3590eeeee25d666f8a5 in /home/pwaller/.local/src/v/cache/vcswork/0cd516681a807dc016da5b7f0387440ea175d644132e9052f6e70a1cbc7e9435: exit status 1
vgo: stat github.com/docker/libnetwork@45b40861e677e37cf27bc184eca5af92f8cdd32d: git fetch --depth=1 origin 45b40861e677e37cf27bc184eca5af92f8cdd32d in /home/pwaller/.local/src/v/cache/vcswork/b336f1f9b64be3a88933393f9543b3e76e4a5bce0bb4b4ed42f9a075055e2a6e: exit status 1
vgo: stat github.com/docker/go-events@18b43f1bc85d9cdd42c05a6cd2d444c7a200a894: git fetch --depth=1 origin 18b43f1bc85d9cdd42c05a6cd2d444c7a200a894 in /home/pwaller/.local/src/v/cache/vcswork/ceca82cd3bc327d14cd6a4b24d913c9734565bc15ec9d417711f0173f1d6dff0: exit status 1
vgo: stat github.com/armon/go-radix@e39d623f12e8e41c7b5529e9a9dd67a1e2261f80: git fetch --depth=1 origin e39d623f12e8e41c7b5529e9a9dd67a1e2261f80 in /home/pwaller/.local/src/v/cache/vcswork/4dc30dad3b06c47b706a6f3ef25202b365a38325f81ffb7cd31c28597072c251: exit status 1
vgo: stat github.com/armon/go-metrics@eb0af217e5e9747e41dd5303755356b62d28e3ec: git fetch --depth=1 origin eb0af217e5e9747e41dd5303755356b62d28e3ec in /home/pwaller/.local/src/v/cache/vcswork/9a9db6e98cab718f3fd21ec86b752de1f36aedfc6d6f10fa4b210ab745a68d0d: exit status 1
vgo: stat github.com/hashicorp/go-msgpack@71c2886f5a673a35f909803f38ece5810165097b: git fetch --depth=1 origin 71c2886f5a673a35f909803f38ece5810165097b in /home/pwaller/.local/src/v/cache/vcswork/84f1269534bbd6f1b0022b2f7f0f7a107c2ecb5fac02cc4a1e234eb3b28be146: exit status 1
vgo: stat github.com/hashicorp/memberlist@88ac4de0d1a0ca6def284b571342db3b777a4c37: git fetch --depth=1 origin 88ac4de0d1a0ca6def284b571342db3b777a4c37 in /home/pwaller/.local/src/v/cache/vcswork/b007b82ebe8d894838e36d700a6a34225b314f02e13d37c50fb2f5e538719c90: exit status 1
vgo: stat github.com/hashicorp/go-multierror@fcdddc395df1ddf4247c69bd436e84cfa0733f7e: git fetch --depth=1 origin fcdddc395df1ddf4247c69bd436e84cfa0733f7e in /home/pwaller/.local/src/v/cache/vcswork/e8986b75deaa1a5c6487ae226bf0521b5bbe482eacbd98923c618b25f3e6d801: exit status 1
vgo: stat github.com/hashicorp/serf@598c54895cc5a7b1a24a398d635e8c0ea0959870: git fetch --depth=1 origin 598c54895cc5a7b1a24a398d635e8c0ea0959870 in /home/pwaller/.local/src/v/cache/vcswork/3466caace3c4c63e70c6a6e75034d611f89475ec5733ebb0dd7c10a6847662f9: exit status 1
vgo: stat github.com/docker/libkv@1d8431073ae03cdaedb198a89722f3aab6d418ef: git fetch --depth=1 origin 1d8431073ae03cdaedb198a89722f3aab6d418ef in /home/pwaller/.local/src/v/cache/vcswork/8cf13d962f8420301cee1b5abcd347b4a05a814c506024123c43a12395242d7d: exit status 1
vgo: stat github.com/vishvananda/netns@604eaf189ee867d8c147fafc28def2394e878d25: git fetch --depth=1 origin 604eaf189ee867d8c147fafc28def2394e878d25 in /home/pwaller/.local/src/v/cache/vcswork/1b270e5db6fb870d6d30b75f73809eaff3df9ef6c030f012d7a9a932b0fb820e: exit status 1
vgo: stat github.com/vishvananda/netlink@482f7a52b758233521878cb6c5904b6bd63f3457: git fetch --depth=1 origin 482f7a52b758233521878cb6c5904b6bd63f3457 in /home/pwaller/.local/src/v/cache/vcswork/4ff59ae4f2cec838b4e8f5ba43b230f5428d35f58a9298babb0097f083d77331: exit status 1
vgo: stat github.com/BurntSushi/toml@f706d00e3de6abe700c994cdd545a1a4915af060: git fetch --depth=1 origin f706d00e3de6abe700c994cdd545a1a4915af060 in /home/pwaller/.local/src/v/cache/vcswork/3c854d7dd8a65b0485436a210dc0dd7f98e1fe4cc8dc0bd3d77527fedda57561: exit status 1
vgo: stat github.com/samuel/go-zookeeper@d0e0d8e11f318e000a8cc434616d69e329edc374: git fetch --depth=1 origin d0e0d8e11f318e000a8cc434616d69e329edc374 in /home/pwaller/.local/src/v/cache/vcswork/f2823631c4fe692d4b0c85da7f20aba66ff731af6c4f40cf7cc90572ce19b48b: exit status 1
vgo: stat github.com/deckarep/golang-set@ef32fa3046d9f249d399f98ebaf9be944430fd1d: git fetch --depth=1 origin ef32fa3046d9f249d399f98ebaf9be944430fd1d in /home/pwaller/.local/src/v/cache/vcswork/2fcf0ce6ea4ec413db0b07090601d4f1abf062d73506e9c0ee0c5aca56f3b2c0: exit status 1
vgo: stat github.com/coreos/etcd@3a49cbb769ebd8d1dd25abb1e83386e9883a5707: git fetch --depth=1 origin 3a49cbb769ebd8d1dd25abb1e83386e9883a5707 in /home/pwaller/.local/src/v/cache/vcswork/e0603d6b678c67453e2d85ba58d5dd0030d8a3ed443832680252eb8f82272b52: exit status 1
vgo: stat github.com/ugorji/go@f1f1a805ed361a0e078bb537e4ea78cd37dcf065: git fetch --depth=1 origin f1f1a805ed361a0e078bb537e4ea78cd37dcf065 in /home/pwaller/.local/src/v/cache/vcswork/4c4ce012b2736486e99bf427f8d15912b07cd69667f52e5aac2af9a9f3c09a9e: exit status 1
vgo: stat github.com/boltdb/bolt@fff57c100f4dea1905678da7e90d92429dff2904: git fetch --depth=1 origin fff57c100f4dea1905678da7e90d92429dff2904 in /home/pwaller/.local/src/v/cache/vcswork/e83c93b3c716fae8fb489742637e698e1bf98745aae3b5226ba355057bdf3717: exit status 1
vgo: stat github.com/miekg/dns@75e6e86cc601825c5dbcd4e0c209eab180997cd7: git fetch --depth=1 origin 75e6e86cc601825c5dbcd4e0c209eab180997cd7 in /home/pwaller/.local/src/v/cache/vcswork/8398ef9e2385ed1171ebb49f31e4200d019135b91a0387e3cc8af4c7405500d0: exit status 1
vgo: stat github.com/mistifyio/go-zfs@22c9b32c84eb0d0c6f4043b6e90fc94073de92fa: git fetch --depth=1 origin 22c9b32c84eb0d0c6f4043b6e90fc94073de92fa in /home/pwaller/.local/src/v/cache/vcswork/64d935563e66336d78f39b3af967a8b6d43ba3407a6d7c51b278681363c15934: exit status 1
vgo: stat github.com/miekg/pkcs11@df8ae6ca730422dba20c768ff38ef7d79077a59f: git fetch --depth=1 origin df8ae6ca730422dba20c768ff38ef7d79077a59f in /home/pwaller/.local/src/v/cache/vcswork/1d74d84db901415b73747d07b43f6cd36ac239739cfe9e183a4501b50e94c3ac: exit status 1
vgo: stat github.com/docker/go@v1.5.1-1-1-gbaf439e: unknown revision "v1.5.1-1-1-gbaf439e"
vgo: stat github.com/agl/ed25519@d2b94fd789ea21d12fac1a4443dd3a3f79cda72c: git fetch --depth=1 origin d2b94fd789ea21d12fac1a4443dd3a3f79cda72c in /home/pwaller/.local/src/v/cache/vcswork/0b82478dc41eb662b31914d0f711acabb12bf913cb746feb99c21463b792f537: exit status 1
vgo: stat github.com/opencontainers/runc@9df8b306d01f59d3a8029be411de015b7304dd8f: git fetch --depth=1 origin 9df8b306d01f59d3a8029be411de015b7304dd8f in /home/pwaller/.local/src/v/cache/vcswork/2d235085b4e185f118a443ea632e2c4f8a6d96ef2b7f068acf187ba102f95dc2: exit status 1
vgo: stat github.com/opencontainers/runtime-spec@1c7c27d043c2a5e513a44084d2b10d77d1402b8c: git fetch --depth=1 origin 1c7c27d043c2a5e513a44084d2b10d77d1402b8c in /home/pwaller/.local/src/v/cache/vcswork/b1301a26e1d0794b2915bbc8037ee62703ad97e00033ed3b117a2c67a23233a3: exit status 1
vgo: stat github.com/seccomp/libseccomp-golang@32f571b70023028bd57d9288c20efbcb237f3ce0: git fetch --depth=1 origin 32f571b70023028bd57d9288c20efbcb237f3ce0 in /home/pwaller/.local/src/v/cache/vcswork/ea142ed8c34e9a56ddedd65452487d29c5612d615c17f06d66852712fc41d44c: exit status 1
vgo: stat github.com/syndtr/gocapability@2c00daeb6c3b45114c80ac44119e7b8801fdd852: git fetch --depth=1 origin 2c00daeb6c3b45114c80ac44119e7b8801fdd852 in /home/pwaller/.local/src/v/cache/vcswork/9325046eaf69fe74b7d765df9e31c31ceb3f57013ea20b2db645b2343184287d: exit status 1
vgo: stat github.com/golang/protobuf@1f49d83d9aa00e6ce4fc8258c71cc7786aec968a: git fetch --depth=1 origin 1f49d83d9aa00e6ce4fc8258c71cc7786aec968a in /home/pwaller/.local/src/v/cache/vcswork/a20e27c072e7660590311c08d0311b908b66675cb634b0caf7cbb860e5c3c705: exit status 1
vgo: stat github.com/Graylog2/go-gelf@aab2f594e4585d43468ac57287b0dece9d806883: git fetch --depth=1 origin aab2f594e4585d43468ac57287b0dece9d806883 in /home/pwaller/.local/src/v/cache/vcswork/82a476495e1f36aa99722958b36546c1a303c12ad00024c272cad343952713c1: exit status 1
vgo: stat github.com/philhofer/fwd@899e4efba8eaa1fea74175308f3fae18ff3319fa: git fetch --depth=1 origin 899e4efba8eaa1fea74175308f3fae18ff3319fa in /home/pwaller/.local/src/v/cache/vcswork/bfcf28133a4a7a2d0a48c7903eb65499a0af01fa3e948952b3db83508a285225: exit status 1
vgo: stat github.com/tinylib/msgp@75ee40d2601edf122ef667e2a07d600d4c44490c: git fetch --depth=1 origin 75ee40d2601edf122ef667e2a07d600d4c44490c in /home/pwaller/.local/src/v/cache/vcswork/9a3d10e611f6412589d884667e8073aa31e525ccd368ee5a32cdfc182b06c5fe: exit status 1
vgo: stat github.com/go-ini/ini@060d7da055ba6ec5ea7a31f116332fe5efa04ce0: git fetch --depth=1 origin 060d7da055ba6ec5ea7a31f116332fe5efa04ce0 in /home/pwaller/.local/src/v/cache/vcswork/f63afdd23e3f4a0ba51aa6e624a654abefeae191beca1b19ebd2c902c09ef9d0: exit status 1
vgo: stat github.com/jmespath/go-jmespath@0b12d6b521d83fc7f755e7cfc1b1fbdd35a01a74: git fetch --depth=1 origin 0b12d6b521d83fc7f755e7cfc1b1fbdd35a01a74 in /home/pwaller/.local/src/v/cache/vcswork/7b1106ecb177564b0bc9784f963c6c785e31d09dcd9f08114684d32af620443f: exit status 1
vgo: stat github.com/bsphere/le_go@d3308aafe090956bc89a65f0769f58251a1b4f03: git fetch --depth=1 origin d3308aafe090956bc89a65f0769f58251a1b4f03 in /home/pwaller/.local/src/v/cache/vcswork/b002d01c40606eccf82e5a89ab7c82f40c9d356c1a4bb6c98e391b675cf86f3f: exit status 1
vgo: stat github.com/docker/docker-credential-helpers@f72c04f1d8e71959a6d103f808c50ccbad79b9fd: git fetch --depth=1 origin f72c04f1d8e71959a6d103f808c50ccbad79b9fd in /home/pwaller/.local/src/v/cache/vcswork/7fc40242592cbce39d68183ed8a6d136a4e2256fc7e88129379f00cc51efb973: exit status 1
vgo: stat github.com/docker/containerd@aa8187dbd3b7ad67d8e5e3a15115d3eef43a7ed1: git fetch --depth=1 origin aa8187dbd3b7ad67d8e5e3a15115d3eef43a7ed1 in /home/pwaller/.local/src/v/cache/vcswork/866a8ef98afb4ede91b1d035c9311173005026d9f972ccab2f432f414987e512: exit status 1
vgo: stat github.com/tonistiigi/fifo@1405643975692217d6720f8b54aeee1bf2cd5cf4: git fetch --depth=1 origin 1405643975692217d6720f8b54aeee1bf2cd5cf4 in /home/pwaller/.local/src/v/cache/vcswork/d9e20cef2b61cf0118375cccb7ca26e0d38ed6b3105b5c8de07976344105b60f: exit status 1
vgo: stat github.com/golang/mock@bd3c8e81be01eef76d4b503f5e687d2d1354d2d9: git fetch --depth=1 origin bd3c8e81be01eef76d4b503f5e687d2d1354d2d9 in /home/pwaller/.local/src/v/cache/vcswork/8fd2fb7f4befaedbf6e7d4c768d11fbe23bc1ce0d138e0689f91529c1898d9a1: exit status 1
vgo: stat github.com/cloudflare/cfssl@7fb22c8cba7ecaf98e4082d22d65800cf45e042a: git fetch --depth=1 origin 7fb22c8cba7ecaf98e4082d22d65800cf45e042a in /home/pwaller/.local/src/v/cache/vcswork/42a0dfebf1eff66b7c2d27c6fb5d7b2b0d6a8f266bbc7dd4fdbd5ac090201bf4: exit status 1
vgo: stat github.com/google/certificate-transparency@d90e65c3a07988180c5b1ece71791c0b6506826e: git fetch --depth=1 origin d90e65c3a07988180c5b1ece71791c0b6506826e in /home/pwaller/.local/src/v/cache/vcswork/91096961f6e56a56781818ef09bd9b5dabc8f04840ba433504b437194b01d65e: exit status 1
vgo: stat github.com/mreiferson/go-httpclient@63fe23f7434723dc904c901043af07931f293c47: git fetch --depth=1 origin 63fe23f7434723dc904c901043af07931f293c47 in /home/pwaller/.local/src/v/cache/vcswork/c2ec4676aae9fa411b819b9ed36e1e1bf7d5c88769b98537683db5205873675e: exit status 1
vgo: stat github.com/hashicorp/go-memdb@608dda3b1410a73eaf3ac8b517c9ae7ebab6aa87: git fetch --depth=1 origin 608dda3b1410a73eaf3ac8b517c9ae7ebab6aa87 in /home/pwaller/.local/src/v/cache/vcswork/242972787ae344da6341fa74d994440ce60c0ef2891b842d908c4c89ea9df05f: exit status 1
vgo: stat github.com/hashicorp/go-immutable-radix@8e8ed81f8f0bf1bdd829593fdd5c29922c1ea990: git fetch --depth=1 origin 8e8ed81f8f0bf1bdd829593fdd5c29922c1ea990 in /home/pwaller/.local/src/v/cache/vcswork/113258058091512d553dab74611129264ab512acc89e82d3e3cec1c07921f25e: exit status 1
vgo: stat github.com/hashicorp/golang-lru@a0d98a5f288019575c6d1f4bb1573fef2d1fcdc4: git fetch --depth=1 origin a0d98a5f288019575c6d1f4bb1573fef2d1fcdc4 in /home/pwaller/.local/src/v/cache/vcswork/1eaaa73a95627cc6e520c43f6cce496386f24ff1b7fa8af090724d7660d929e9: exit status 1
vgo: stat github.com/coreos/pkg@fa29b1d70f0beaddd4c7021607cc3c3be8ce94b8: git fetch --depth=1 origin fa29b1d70f0beaddd4c7021607cc3c3be8ce94b8 in /home/pwaller/.local/src/v/cache/vcswork/f7c8ebd7bb3e4073cf5802eb968d76a46c1791afe1a1135369d2f934825d2a0b: exit status 1
vgo: stat github.com/pivotal-golang/clock@3fd3c1944c59d9742e1cd333672181cd1a6f9fa0: git fetch --depth=1 origin 3fd3c1944c59d9742e1cd333672181cd1a6f9fa0 in /home/pwaller/.local/src/v/cache/vcswork/327a10928fdcae61e5fd5a109282c83f1a2bec4719a08ad511de884480edd964: exit status 1
vgo: stat github.com/prometheus/client_golang@52437c81da6b127a9925d17eb3a382a2e5fd395e: git fetch --depth=1 origin 52437c81da6b127a9925d17eb3a382a2e5fd395e in /home/pwaller/.local/src/v/cache/vcswork/3785b9359e0a97a9dc849585696cf17189b86b713b7e59c2842fb3c44d3ff610: exit status 1
vgo: stat github.com/beorn7/perks@4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9: git fetch --depth=1 origin 4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9 in /home/pwaller/.local/src/v/cache/vcswork/1a437ed9cd95ef0b3850ad46f66cacade14a3fca4584dad97c8e5299ff7b4735: exit status 1
vgo: stat github.com/prometheus/client_model@fa8ad6fec33561be4280a8f0514318c79d7f6cb6: git fetch --depth=1 origin fa8ad6fec33561be4280a8f0514318c79d7f6cb6 in /home/pwaller/.local/src/v/cache/vcswork/2a98e665081184f4ca01f0af8738c882495d1fb131b7ed20ad844d3ba1bb6393: exit status 1
vgo: stat github.com/prometheus/common@ebdfc6da46522d58825777cf1f90490a5b1ef1d8: git fetch --depth=1 origin ebdfc6da46522d58825777cf1f90490a5b1ef1d8 in /home/pwaller/.local/src/v/cache/vcswork/78eeb4629eb558fa221be26a69bd8019d99c56c9d5e61a056019c7d4845bf714: exit status 1
vgo: stat github.com/prometheus/procfs@abf152e5f3e97f2fafac028d2cc06c1feb87ffa5: git fetch --depth=1 origin abf152e5f3e97f2fafac028d2cc06c1feb87ffa5 in /home/pwaller/.local/src/v/cache/vcswork/98c949bd9adf825ba7df08da3b4256328548aa917c603094bfc20a4042505d6d: exit status 1
vgo: stat bitbucket.org/ww/goautoneg@75cd24fc2f2c2a2088577d12123ddee5f54e0675: git fetch --depth=1 origin 75cd24fc2f2c2a2088577d12123ddee5f54e0675 in /home/pwaller/.local/src/v/cache/vcswork/52eb1c0593bdece4d4ec1f51b6d400c08ce21540252af4c262ec3a940587d7ee: exit status 128:
	remote: Not Found
	fatal: unable to access 'https://bitbucket.org/ww/goautoneg/': GnuTLS recv error (-110): The TLS connection was non-properly terminated.
vgo: stat github.com/matttproud/golang_protobuf_extensions@fc2b8d3a73c4867e51861bbdd5ae3c1f0869dd6a: git fetch --depth=1 origin fc2b8d3a73c4867e51861bbdd5ae3c1f0869dd6a in /home/pwaller/.local/src/v/cache/vcswork/c9f9bbb8cc68e928d83601198e57e7c65ab57e6945016296f48e56d8e0f8d013: exit status 1
vgo: stat github.com/pkg/errors@839d9e913e063e28dfd0e6c7b7512793e0a48be9: git fetch --depth=1 origin 839d9e913e063e28dfd0e6c7b7512793e0a48be9 in /home/pwaller/.local/src/v/cache/vcswork/9b57de15915a2564a133192909d2d779433a38d49df7d581dc764e6764a41406: exit status 1
vgo: stat github.com/spf13/cobra@v1.5: unknown revision "v1.5"
vgo: stat github.com/spf13/pflag@dabebe21bf790f782ea4c7bbd2efc430de182afd: git fetch --depth=1 origin dabebe21bf790f782ea4c7bbd2efc430de182afd in /home/pwaller/.local/src/v/cache/vcswork/389cbbf79b0218a16e6f902e349b1cabca23e0203c06f228d24031e72b6cf480: exit status 1
vgo: stat github.com/docker/go-metrics@86138d05f285fd9737a99bee2d9be30866b59d72: git fetch --depth=1 origin 86138d05f285fd9737a99bee2d9be30866b59d72 in /home/pwaller/.local/src/v/cache/vcswork/b5cbc01bb6eaa02d55f163ff45f255a26c3ed0efd808299516c8f2ead92c5d37: exit status 1
vgo: stat github.com/mitchellh/mapstructure@f3009df150dadf309fdee4a54ed65c124afad715: git fetch --depth=1 origin f3009df150dadf309fdee4a54ed65c124afad715 in /home/pwaller/.local/src/v/cache/vcswork/cbe18bc04d35c650864e89ab9c27c6de5e24d51255fad0f64b974d13d4f47330: exit status 1
vgo: stat github.com/xeipuuv/gojsonpointer@e0fe6f68307607d540ed8eac07a342c33fa1b54a: git fetch --depth=1 origin e0fe6f68307607d540ed8eac07a342c33fa1b54a in /home/pwaller/.local/src/v/cache/vcswork/9edb38f89cde3f8c5a3287be807e2aec4d57ae79713260fde4ffac2f27561fa3: exit status 1
vgo: stat github.com/xeipuuv/gojsonreference@e02fc20de94c78484cd5ffb007f8af96be030a45: git fetch --depth=1 origin e02fc20de94c78484cd5ffb007f8af96be030a45 in /home/pwaller/.local/src/v/cache/vcswork/1c853cf6997cc6c35b7186645cec467477abe75023b90852026434182d97eee7: exit status 1
vgo: stat github.com/xeipuuv/gojsonschema@93e72a773fade158921402d6a24c819b48aba29d: git fetch --depth=1 origin 93e72a773fade158921402d6a24c819b48aba29d in /home/pwaller/.local/src/v/cache/vcswork/909711cbd0befab6d0fcb0681b420602d9150b9e3be795db9f9715926f08a149: exit status 1
vgo: finding github.com/docker/docker (latest)
vgo: adding github.com/docker/docker v1.13.1
vgo: finding github.com/docker/docker v1.13.1
vgo: stat github.com/Azure/go-ansiterm@388960b655244e76e24c75f48631564eaefade62: git fetch --depth=1 https://github.com/Azure/go-ansiterm 388960b655244e76e24c75f48631564eaefade62 in /home/pwaller/.local/src/v/cache/vcswork/da5145fda272732cd74527dacbe4967cc6a648dad514ccb83c288a28ea4c0671: exit status 1
vgo: stat github.com/davecgh/go-spew@6d212800a42e8ab5c146b8ace3490ee17e5225f9: git fetch --depth=1 https://github.com/davecgh/go-spew 6d212800a42e8ab5c146b8ace3490ee17e5225f9 in /home/pwaller/.local/src/v/cache/vcswork/b9a4b9bbdb4a59723f2348415ad7ffda91568455a1cfd92e97976132bdfbaf57: exit status 1
vgo: stat github.com/docker/libtrust@9cbd2a1374f46905c68a4eb3694a130610adc62a: git fetch --depth=1 https://github.com/docker/libtrust 9cbd2a1374f46905c68a4eb3694a130610adc62a in /home/pwaller/.local/src/v/cache/vcswork/da07fa0f90047c1906572bf6085c832cc0d3e6a5b251b2c1daacbbe31dba899f: exit status 1
vgo: stat github.com/go-check/check@4ed411733c5785b40214c70bce814c3a3a689609: git fetch --depth=1 https://github.com/go-check/check 4ed411733c5785b40214c70bce814c3a3a689609 in /home/pwaller/.local/src/v/cache/vcswork/8a10bcc540d5cc49b62e7151b7bf7502646412570399a1a93fbc8c13a5923bf0: exit status 1
vgo: stat golang.org/x/net@2beffdc2e92c8a3027590f898fe88f69af48a3f8: git fetch --depth=1 https://go.googlesource.com/net 2beffdc2e92c8a3027590f898fe88f69af48a3f8 in /home/pwaller/.local/src/v/cache/vcswork/4a22365141bc4eea5d5ac4a1395e653f2669485db75ef119e7bbec8e19b12a21: exit status 128:
	fatal: expected shallow/unshallow, got ERR internal server error
	fatal: The remote end hung up unexpectedly
vgo: stat github.com/docker/go-units@8a7beacffa3009a9ac66bad506b18ffdd110cf97: git fetch --depth=1 https://github.com/docker/go-units 8a7beacffa3009a9ac66bad506b18ffdd110cf97 in /home/pwaller/.local/src/v/cache/vcswork/9f2d1a527162210923891d0d8f65d59f937c6335cd9cad946f750ba191817910: exit status 1
vgo: stat github.com/RackSec/srslog@456df3a81436d29ba874f3590eeeee25d666f8a5: git fetch --depth=1 https://github.com/RackSec/srslog 456df3a81436d29ba874f3590eeeee25d666f8a5 in /home/pwaller/.local/src/v/cache/vcswork/0cd516681a807dc016da5b7f0387440ea175d644132e9052f6e70a1cbc7e9435: exit status 1
vgo: stat github.com/docker/libnetwork@45b40861e677e37cf27bc184eca5af92f8cdd32d: git fetch --depth=1 https://github.com/docker/libnetwork 45b40861e677e37cf27bc184eca5af92f8cdd32d in /home/pwaller/.local/src/v/cache/vcswork/b336f1f9b64be3a88933393f9543b3e76e4a5bce0bb4b4ed42f9a075055e2a6e: exit status 1
vgo: stat github.com/docker/go-events@18b43f1bc85d9cdd42c05a6cd2d444c7a200a894: git fetch --depth=1 https://github.com/docker/go-events 18b43f1bc85d9cdd42c05a6cd2d444c7a200a894 in /home/pwaller/.local/src/v/cache/vcswork/ceca82cd3bc327d14cd6a4b24d913c9734565bc15ec9d417711f0173f1d6dff0: exit status 1
vgo: stat github.com/armon/go-radix@e39d623f12e8e41c7b5529e9a9dd67a1e2261f80: git fetch --depth=1 https://github.com/armon/go-radix e39d623f12e8e41c7b5529e9a9dd67a1e2261f80 in /home/pwaller/.local/src/v/cache/vcswork/4dc30dad3b06c47b706a6f3ef25202b365a38325f81ffb7cd31c28597072c251: exit status 1
vgo: stat github.com/armon/go-metrics@eb0af217e5e9747e41dd5303755356b62d28e3ec: git fetch --depth=1 https://github.com/armon/go-metrics eb0af217e5e9747e41dd5303755356b62d28e3ec in /home/pwaller/.local/src/v/cache/vcswork/9a9db6e98cab718f3fd21ec86b752de1f36aedfc6d6f10fa4b210ab745a68d0d: exit status 1
vgo: stat github.com/hashicorp/go-msgpack@71c2886f5a673a35f909803f38ece5810165097b: git fetch --depth=1 https://github.com/hashicorp/go-msgpack 71c2886f5a673a35f909803f38ece5810165097b in /home/pwaller/.local/src/v/cache/vcswork/84f1269534bbd6f1b0022b2f7f0f7a107c2ecb5fac02cc4a1e234eb3b28be146: exit status 1
vgo: stat github.com/hashicorp/memberlist@88ac4de0d1a0ca6def284b571342db3b777a4c37: git fetch --depth=1 https://github.com/hashicorp/memberlist 88ac4de0d1a0ca6def284b571342db3b777a4c37 in /home/pwaller/.local/src/v/cache/vcswork/b007b82ebe8d894838e36d700a6a34225b314f02e13d37c50fb2f5e538719c90: exit status 1
vgo: stat github.com/hashicorp/go-multierror@fcdddc395df1ddf4247c69bd436e84cfa0733f7e: git fetch --depth=1 https://github.com/hashicorp/go-multierror fcdddc395df1ddf4247c69bd436e84cfa0733f7e in /home/pwaller/.local/src/v/cache/vcswork/e8986b75deaa1a5c6487ae226bf0521b5bbe482eacbd98923c618b25f3e6d801: exit status 1
vgo: stat github.com/hashicorp/serf@598c54895cc5a7b1a24a398d635e8c0ea0959870: git fetch --depth=1 https://github.com/hashicorp/serf 598c54895cc5a7b1a24a398d635e8c0ea0959870 in /home/pwaller/.local/src/v/cache/vcswork/3466caace3c4c63e70c6a6e75034d611f89475ec5733ebb0dd7c10a6847662f9: exit status 1
vgo: stat github.com/docker/libkv@1d8431073ae03cdaedb198a89722f3aab6d418ef: git fetch --depth=1 https://github.com/docker/libkv 1d8431073ae03cdaedb198a89722f3aab6d418ef in /home/pwaller/.local/src/v/cache/vcswork/8cf13d962f8420301cee1b5abcd347b4a05a814c506024123c43a12395242d7d: exit status 1
vgo: stat github.com/vishvananda/netns@604eaf189ee867d8c147fafc28def2394e878d25: git fetch --depth=1 https://github.com/vishvananda/netns 604eaf189ee867d8c147fafc28def2394e878d25 in /home/pwaller/.local/src/v/cache/vcswork/1b270e5db6fb870d6d30b75f73809eaff3df9ef6c030f012d7a9a932b0fb820e: exit status 1
vgo: stat github.com/vishvananda/netlink@482f7a52b758233521878cb6c5904b6bd63f3457: git fetch --depth=1 https://github.com/vishvananda/netlink 482f7a52b758233521878cb6c5904b6bd63f3457 in /home/pwaller/.local/src/v/cache/vcswork/4ff59ae4f2cec838b4e8f5ba43b230f5428d35f58a9298babb0097f083d77331: exit status 1
vgo: stat github.com/BurntSushi/toml@f706d00e3de6abe700c994cdd545a1a4915af060: git fetch --depth=1 https://github.com/BurntSushi/toml f706d00e3de6abe700c994cdd545a1a4915af060 in /home/pwaller/.local/src/v/cache/vcswork/3c854d7dd8a65b0485436a210dc0dd7f98e1fe4cc8dc0bd3d77527fedda57561: exit status 1
vgo: stat github.com/samuel/go-zookeeper@d0e0d8e11f318e000a8cc434616d69e329edc374: git fetch --depth=1 https://github.com/samuel/go-zookeeper d0e0d8e11f318e000a8cc434616d69e329edc374 in /home/pwaller/.local/src/v/cache/vcswork/f2823631c4fe692d4b0c85da7f20aba66ff731af6c4f40cf7cc90572ce19b48b: exit status 1
vgo: stat github.com/deckarep/golang-set@ef32fa3046d9f249d399f98ebaf9be944430fd1d: git fetch --depth=1 https://github.com/deckarep/golang-set ef32fa3046d9f249d399f98ebaf9be944430fd1d in /home/pwaller/.local/src/v/cache/vcswork/2fcf0ce6ea4ec413db0b07090601d4f1abf062d73506e9c0ee0c5aca56f3b2c0: exit status 1
vgo: stat github.com/coreos/etcd@3a49cbb769ebd8d1dd25abb1e83386e9883a5707: git fetch --depth=1 https://github.com/coreos/etcd 3a49cbb769ebd8d1dd25abb1e83386e9883a5707 in /home/pwaller/.local/src/v/cache/vcswork/e0603d6b678c67453e2d85ba58d5dd0030d8a3ed443832680252eb8f82272b52: exit status 1
vgo: stat github.com/ugorji/go@f1f1a805ed361a0e078bb537e4ea78cd37dcf065: git fetch --depth=1 https://github.com/ugorji/go f1f1a805ed361a0e078bb537e4ea78cd37dcf065 in /home/pwaller/.local/src/v/cache/vcswork/4c4ce012b2736486e99bf427f8d15912b07cd69667f52e5aac2af9a9f3c09a9e: exit status 1
vgo: stat github.com/boltdb/bolt@fff57c100f4dea1905678da7e90d92429dff2904: git fetch --depth=1 https://github.com/boltdb/bolt fff57c100f4dea1905678da7e90d92429dff2904 in /home/pwaller/.local/src/v/cache/vcswork/e83c93b3c716fae8fb489742637e698e1bf98745aae3b5226ba355057bdf3717: exit status 1
vgo: stat github.com/miekg/dns@75e6e86cc601825c5dbcd4e0c209eab180997cd7: git fetch --depth=1 https://github.com/miekg/dns 75e6e86cc601825c5dbcd4e0c209eab180997cd7 in /home/pwaller/.local/src/v/cache/vcswork/8398ef9e2385ed1171ebb49f31e4200d019135b91a0387e3cc8af4c7405500d0: exit status 1
vgo: stat github.com/mistifyio/go-zfs@22c9b32c84eb0d0c6f4043b6e90fc94073de92fa: git fetch --depth=1 https://github.com/mistifyio/go-zfs 22c9b32c84eb0d0c6f4043b6e90fc94073de92fa in /home/pwaller/.local/src/v/cache/vcswork/64d935563e66336d78f39b3af967a8b6d43ba3407a6d7c51b278681363c15934: exit status 1
vgo: stat github.com/miekg/pkcs11@df8ae6ca730422dba20c768ff38ef7d79077a59f: git fetch --depth=1 https://github.com/miekg/pkcs11 df8ae6ca730422dba20c768ff38ef7d79077a59f in /home/pwaller/.local/src/v/cache/vcswork/1d74d84db901415b73747d07b43f6cd36ac239739cfe9e183a4501b50e94c3ac: exit status 1
vgo: stat github.com/docker/go@v1.5.1-1-1-gbaf439e: unknown revision "v1.5.1-1-1-gbaf439e"
vgo: stat github.com/agl/ed25519@d2b94fd789ea21d12fac1a4443dd3a3f79cda72c: git fetch --depth=1 https://github.com/agl/ed25519 d2b94fd789ea21d12fac1a4443dd3a3f79cda72c in /home/pwaller/.local/src/v/cache/vcswork/0b82478dc41eb662b31914d0f711acabb12bf913cb746feb99c21463b792f537: exit status 1
vgo: stat github.com/opencontainers/runc@9df8b306d01f59d3a8029be411de015b7304dd8f: git fetch --depth=1 https://github.com/opencontainers/runc 9df8b306d01f59d3a8029be411de015b7304dd8f in /home/pwaller/.local/src/v/cache/vcswork/2d235085b4e185f118a443ea632e2c4f8a6d96ef2b7f068acf187ba102f95dc2: exit status 1
vgo: stat github.com/opencontainers/runtime-spec@1c7c27d043c2a5e513a44084d2b10d77d1402b8c: git fetch --depth=1 https://github.com/opencontainers/runtime-spec 1c7c27d043c2a5e513a44084d2b10d77d1402b8c in /home/pwaller/.local/src/v/cache/vcswork/b1301a26e1d0794b2915bbc8037ee62703ad97e00033ed3b117a2c67a23233a3: exit status 1
vgo: stat github.com/seccomp/libseccomp-golang@32f571b70023028bd57d9288c20efbcb237f3ce0: git fetch --depth=1 https://github.com/seccomp/libseccomp-golang 32f571b70023028bd57d9288c20efbcb237f3ce0 in /home/pwaller/.local/src/v/cache/vcswork/ea142ed8c34e9a56ddedd65452487d29c5612d615c17f06d66852712fc41d44c: exit status 1
vgo: stat github.com/syndtr/gocapability@2c00daeb6c3b45114c80ac44119e7b8801fdd852: git fetch --depth=1 https://github.com/syndtr/gocapability 2c00daeb6c3b45114c80ac44119e7b8801fdd852 in /home/pwaller/.local/src/v/cache/vcswork/9325046eaf69fe74b7d765df9e31c31ceb3f57013ea20b2db645b2343184287d: exit status 1
vgo: stat github.com/golang/protobuf@1f49d83d9aa00e6ce4fc8258c71cc7786aec968a: git fetch --depth=1 https://github.com/golang/protobuf 1f49d83d9aa00e6ce4fc8258c71cc7786aec968a in /home/pwaller/.local/src/v/cache/vcswork/a20e27c072e7660590311c08d0311b908b66675cb634b0caf7cbb860e5c3c705: exit status 1
vgo: stat github.com/Graylog2/go-gelf@aab2f594e4585d43468ac57287b0dece9d806883: git fetch --depth=1 https://github.com/Graylog2/go-gelf aab2f594e4585d43468ac57287b0dece9d806883 in /home/pwaller/.local/src/v/cache/vcswork/82a476495e1f36aa99722958b36546c1a303c12ad00024c272cad343952713c1: exit status 1
vgo: stat github.com/philhofer/fwd@899e4efba8eaa1fea74175308f3fae18ff3319fa: git fetch --depth=1 https://github.com/philhofer/fwd 899e4efba8eaa1fea74175308f3fae18ff3319fa in /home/pwaller/.local/src/v/cache/vcswork/bfcf28133a4a7a2d0a48c7903eb65499a0af01fa3e948952b3db83508a285225: exit status 1
vgo: stat github.com/tinylib/msgp@75ee40d2601edf122ef667e2a07d600d4c44490c: git fetch --depth=1 https://github.com/tinylib/msgp 75ee40d2601edf122ef667e2a07d600d4c44490c in /home/pwaller/.local/src/v/cache/vcswork/9a3d10e611f6412589d884667e8073aa31e525ccd368ee5a32cdfc182b06c5fe: exit status 1
vgo: stat github.com/go-ini/ini@060d7da055ba6ec5ea7a31f116332fe5efa04ce0: git fetch --depth=1 https://github.com/go-ini/ini 060d7da055ba6ec5ea7a31f116332fe5efa04ce0 in /home/pwaller/.local/src/v/cache/vcswork/f63afdd23e3f4a0ba51aa6e624a654abefeae191beca1b19ebd2c902c09ef9d0: exit status 1
vgo: stat github.com/jmespath/go-jmespath@0b12d6b521d83fc7f755e7cfc1b1fbdd35a01a74: git fetch --depth=1 https://github.com/jmespath/go-jmespath 0b12d6b521d83fc7f755e7cfc1b1fbdd35a01a74 in /home/pwaller/.local/src/v/cache/vcswork/7b1106ecb177564b0bc9784f963c6c785e31d09dcd9f08114684d32af620443f: exit status 1
vgo: stat github.com/bsphere/le_go@d3308aafe090956bc89a65f0769f58251a1b4f03: git fetch --depth=1 https://github.com/bsphere/le_go d3308aafe090956bc89a65f0769f58251a1b4f03 in /home/pwaller/.local/src/v/cache/vcswork/b002d01c40606eccf82e5a89ab7c82f40c9d356c1a4bb6c98e391b675cf86f3f: exit status 1
vgo: stat github.com/docker/docker-credential-helpers@f72c04f1d8e71959a6d103f808c50ccbad79b9fd: git fetch --depth=1 https://github.com/docker/docker-credential-helpers f72c04f1d8e71959a6d103f808c50ccbad79b9fd in /home/pwaller/.local/src/v/cache/vcswork/7fc40242592cbce39d68183ed8a6d136a4e2256fc7e88129379f00cc51efb973: exit status 1
vgo: stat github.com/docker/containerd@aa8187dbd3b7ad67d8e5e3a15115d3eef43a7ed1: git fetch --depth=1 https://github.com/docker/containerd aa8187dbd3b7ad67d8e5e3a15115d3eef43a7ed1 in /home/pwaller/.local/src/v/cache/vcswork/866a8ef98afb4ede91b1d035c9311173005026d9f972ccab2f432f414987e512: exit status 1
vgo: stat github.com/tonistiigi/fifo@1405643975692217d6720f8b54aeee1bf2cd5cf4: git fetch --depth=1 https://github.com/tonistiigi/fifo 1405643975692217d6720f8b54aeee1bf2cd5cf4 in /home/pwaller/.local/src/v/cache/vcswork/d9e20cef2b61cf0118375cccb7ca26e0d38ed6b3105b5c8de07976344105b60f: exit status 1
vgo: stat github.com/golang/mock@bd3c8e81be01eef76d4b503f5e687d2d1354d2d9: git fetch --depth=1 https://github.com/golang/mock bd3c8e81be01eef76d4b503f5e687d2d1354d2d9 in /home/pwaller/.local/src/v/cache/vcswork/8fd2fb7f4befaedbf6e7d4c768d11fbe23bc1ce0d138e0689f91529c1898d9a1: exit status 1
vgo: stat github.com/cloudflare/cfssl@7fb22c8cba7ecaf98e4082d22d65800cf45e042a: git fetch --depth=1 https://github.com/cloudflare/cfssl 7fb22c8cba7ecaf98e4082d22d65800cf45e042a in /home/pwaller/.local/src/v/cache/vcswork/42a0dfebf1eff66b7c2d27c6fb5d7b2b0d6a8f266bbc7dd4fdbd5ac090201bf4: exit status 1
vgo: stat github.com/google/certificate-transparency@d90e65c3a07988180c5b1ece71791c0b6506826e: git fetch --depth=1 https://github.com/google/certificate-transparency d90e65c3a07988180c5b1ece71791c0b6506826e in /home/pwaller/.local/src/v/cache/vcswork/91096961f6e56a56781818ef09bd9b5dabc8f04840ba433504b437194b01d65e: exit status 1
vgo: stat github.com/mreiferson/go-httpclient@63fe23f7434723dc904c901043af07931f293c47: git fetch --depth=1 https://github.com/mreiferson/go-httpclient 63fe23f7434723dc904c901043af07931f293c47 in /home/pwaller/.local/src/v/cache/vcswork/c2ec4676aae9fa411b819b9ed36e1e1bf7d5c88769b98537683db5205873675e: exit status 1
vgo: stat github.com/hashicorp/go-memdb@608dda3b1410a73eaf3ac8b517c9ae7ebab6aa87: git fetch --depth=1 https://github.com/hashicorp/go-memdb 608dda3b1410a73eaf3ac8b517c9ae7ebab6aa87 in /home/pwaller/.local/src/v/cache/vcswork/242972787ae344da6341fa74d994440ce60c0ef2891b842d908c4c89ea9df05f: exit status 1
vgo: stat github.com/hashicorp/go-immutable-radix@8e8ed81f8f0bf1bdd829593fdd5c29922c1ea990: git fetch --depth=1 https://github.com/hashicorp/go-immutable-radix 8e8ed81f8f0bf1bdd829593fdd5c29922c1ea990 in /home/pwaller/.local/src/v/cache/vcswork/113258058091512d553dab74611129264ab512acc89e82d3e3cec1c07921f25e: exit status 1
vgo: stat github.com/hashicorp/golang-lru@a0d98a5f288019575c6d1f4bb1573fef2d1fcdc4: git fetch --depth=1 https://github.com/hashicorp/golang-lru a0d98a5f288019575c6d1f4bb1573fef2d1fcdc4 in /home/pwaller/.local/src/v/cache/vcswork/1eaaa73a95627cc6e520c43f6cce496386f24ff1b7fa8af090724d7660d929e9: exit status 1
vgo: stat github.com/coreos/pkg@fa29b1d70f0beaddd4c7021607cc3c3be8ce94b8: git fetch --depth=1 https://github.com/coreos/pkg fa29b1d70f0beaddd4c7021607cc3c3be8ce94b8 in /home/pwaller/.local/src/v/cache/vcswork/f7c8ebd7bb3e4073cf5802eb968d76a46c1791afe1a1135369d2f934825d2a0b: exit status 1
vgo: stat github.com/pivotal-golang/clock@3fd3c1944c59d9742e1cd333672181cd1a6f9fa0: git fetch --depth=1 https://github.com/pivotal-golang/clock 3fd3c1944c59d9742e1cd333672181cd1a6f9fa0 in /home/pwaller/.local/src/v/cache/vcswork/327a10928fdcae61e5fd5a109282c83f1a2bec4719a08ad511de884480edd964: exit status 1
vgo: stat github.com/prometheus/client_golang@52437c81da6b127a9925d17eb3a382a2e5fd395e: git fetch --depth=1 https://github.com/prometheus/client_golang 52437c81da6b127a9925d17eb3a382a2e5fd395e in /home/pwaller/.local/src/v/cache/vcswork/3785b9359e0a97a9dc849585696cf17189b86b713b7e59c2842fb3c44d3ff610: exit status 1
vgo: stat github.com/beorn7/perks@4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9: git fetch --depth=1 https://github.com/beorn7/perks 4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9 in /home/pwaller/.local/src/v/cache/vcswork/1a437ed9cd95ef0b3850ad46f66cacade14a3fca4584dad97c8e5299ff7b4735: exit status 1
vgo: stat github.com/prometheus/client_model@fa8ad6fec33561be4280a8f0514318c79d7f6cb6: git fetch --depth=1 https://github.com/prometheus/client_model fa8ad6fec33561be4280a8f0514318c79d7f6cb6 in /home/pwaller/.local/src/v/cache/vcswork/2a98e665081184f4ca01f0af8738c882495d1fb131b7ed20ad844d3ba1bb6393: exit status 1
vgo: stat github.com/prometheus/common@ebdfc6da46522d58825777cf1f90490a5b1ef1d8: git fetch --depth=1 https://github.com/prometheus/common ebdfc6da46522d58825777cf1f90490a5b1ef1d8 in /home/pwaller/.local/src/v/cache/vcswork/78eeb4629eb558fa221be26a69bd8019d99c56c9d5e61a056019c7d4845bf714: exit status 1
vgo: stat github.com/prometheus/procfs@abf152e5f3e97f2fafac028d2cc06c1feb87ffa5: git fetch --depth=1 https://github.com/prometheus/procfs abf152e5f3e97f2fafac028d2cc06c1feb87ffa5 in /home/pwaller/.local/src/v/cache/vcswork/98c949bd9adf825ba7df08da3b4256328548aa917c603094bfc20a4042505d6d: exit status 1
vgo: stat bitbucket.org/ww/goautoneg@75cd24fc2f2c2a2088577d12123ddee5f54e0675: git fetch --depth=1 https://bitbucket.org/ww/goautoneg 75cd24fc2f2c2a2088577d12123ddee5f54e0675 in /home/pwaller/.local/src/v/cache/vcswork/52eb1c0593bdece4d4ec1f51b6d400c08ce21540252af4c262ec3a940587d7ee: exit status 128:
	remote: Not Found
	fatal: unable to access 'https://bitbucket.org/ww/goautoneg/': GnuTLS recv error (-110): The TLS connection was non-properly terminated.
vgo: stat github.com/matttproud/golang_protobuf_extensions@fc2b8d3a73c4867e51861bbdd5ae3c1f0869dd6a: git fetch --depth=1 https://github.com/matttproud/golang_protobuf_extensions fc2b8d3a73c4867e51861bbdd5ae3c1f0869dd6a in /home/pwaller/.local/src/v/cache/vcswork/c9f9bbb8cc68e928d83601198e57e7c65ab57e6945016296f48e56d8e0f8d013: exit status 1
vgo: stat github.com/pkg/errors@839d9e913e063e28dfd0e6c7b7512793e0a48be9: git fetch --depth=1 https://github.com/pkg/errors 839d9e913e063e28dfd0e6c7b7512793e0a48be9 in /home/pwaller/.local/src/v/cache/vcswork/9b57de15915a2564a133192909d2d779433a38d49df7d581dc764e6764a41406: exit status 1
vgo: stat github.com/spf13/cobra@v1.5: unknown revision "v1.5"
vgo: stat github.com/spf13/pflag@dabebe21bf790f782ea4c7bbd2efc430de182afd: git fetch --depth=1 https://github.com/spf13/pflag dabebe21bf790f782ea4c7bbd2efc430de182afd in /home/pwaller/.local/src/v/cache/vcswork/389cbbf79b0218a16e6f902e349b1cabca23e0203c06f228d24031e72b6cf480: exit status 1
vgo: stat github.com/docker/go-metrics@86138d05f285fd9737a99bee2d9be30866b59d72: git fetch --depth=1 https://github.com/docker/go-metrics 86138d05f285fd9737a99bee2d9be30866b59d72 in /home/pwaller/.local/src/v/cache/vcswork/b5cbc01bb6eaa02d55f163ff45f255a26c3ed0efd808299516c8f2ead92c5d37: exit status 1
vgo: stat github.com/mitchellh/mapstructure@f3009df150dadf309fdee4a54ed65c124afad715: git fetch --depth=1 https://github.com/mitchellh/mapstructure f3009df150dadf309fdee4a54ed65c124afad715 in /home/pwaller/.local/src/v/cache/vcswork/cbe18bc04d35c650864e89ab9c27c6de5e24d51255fad0f64b974d13d4f47330: exit status 1
vgo: stat github.com/xeipuuv/gojsonpointer@e0fe6f68307607d540ed8eac07a342c33fa1b54a: git fetch --depth=1 https://github.com/xeipuuv/gojsonpointer e0fe6f68307607d540ed8eac07a342c33fa1b54a in /home/pwaller/.local/src/v/cache/vcswork/9edb38f89cde3f8c5a3287be807e2aec4d57ae79713260fde4ffac2f27561fa3: exit status 1
vgo: stat github.com/xeipuuv/gojsonreference@e02fc20de94c78484cd5ffb007f8af96be030a45: git fetch --depth=1 https://github.com/xeipuuv/gojsonreference e02fc20de94c78484cd5ffb007f8af96be030a45 in /home/pwaller/.local/src/v/cache/vcswork/1c853cf6997cc6c35b7186645cec467477abe75023b90852026434182d97eee7: exit status 1
vgo: stat github.com/xeipuuv/gojsonschema@93e72a773fade158921402d6a24c819b48aba29d: git fetch --depth=1 https://github.com/xeipuuv/gojsonschema 93e72a773fade158921402d6a24c819b48aba29d in /home/pwaller/.local/src/v/cache/vcswork/909711cbd0befab6d0fcb0681b420602d9150b9e3be795db9f9715926f08a149: exit status 1
vgo: finding gopkg.in/yaml.v2 v2.0.0-20160301204022-a83829b6f129
vgo: finding google.golang.org/grpc v1.0.2
vgo: finding google.golang.org/cloud v0.0.0-20151218002640-dae7e3d993bc
vgo: finding google.golang.org/api v0.0.0-20151217002415-dc6d2353af16
vgo: finding golang.org/x/time v0.0.0-20160202183820-a4bde1265759
vgo: finding golang.org/x/sys v0.0.0-20160916181909-8f0908ab3b24
vgo: finding golang.org/x/oauth2 v0.0.0-20151204193638-2baa8a1b9338
vgo: finding golang.org/x/crypto v0.0.0-20160408163010-3fbbcd23f1cb
vgo: finding github.com/vdemeester/shakers v0.0.0-20160210082636-24d7f1d6a71a
vgo: finding github.com/vbatts/tar-split v0.10.1
vgo: finding github.com/tchap/go-patricia v0.0.0-20160811102535-666120de432a
vgo: finding github.com/pborman/uuid v0.0.0-20160209185913-a97ce2ca70fa
vgo: finding github.com/mattn/go-sqlite3 v1.1.0
vgo: finding github.com/mattn/go-shellwords v1.0.0
vgo: finding github.com/kr/pty v0.0.0-20150511174710-5cf931ef8f76
vgo: finding github.com/inconshreveable/mousetrap v0.0.0-20141017200713-76626ae9c91c
vgo: finding github.com/imdario/mergo v0.0.0-20151231081848-bc0f15622cd2
vgo: finding github.com/hashicorp/consul v0.5.2
vgo: finding github.com/gorilla/mux v0.0.0-20160317213430-0eeaf8392f5b
vgo: finding github.com/gorilla/context v0.0.0-20160226214623-1ea25387ff6f
vgo: finding github.com/gogo/protobuf v0.0.0-20160824171236-909568be09de
vgo: finding github.com/godbus/dbus v0.0.0-20160408155243-5f6efc7ef275
vgo: finding github.com/fsnotify/fsnotify v1.2.11
vgo: finding github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568
vgo: finding github.com/fluent/fluent-logger-golang v1.2.1
vgo: finding github.com/docker/swarmkit v0.0.0-20170201211723-1c7f003d75f0
vgo: stat github.com/golang/protobuf@3c84672111d91bb5ac31719e112f9f7126a0e26e: git fetch --depth=1 https://github.com/golang/protobuf 3c84672111d91bb5ac31719e112f9f7126a0e26e in /home/pwaller/.local/src/v/cache/vcswork/a20e27c072e7660590311c08d0311b908b66675cb634b0caf7cbb860e5c3c705: exit status 1
vgo: stat github.com/coreos/etcd@bd7107bd4bf26219ba9852aa6c4c817ccde0191c: git fetch --depth=1 https://github.com/coreos/etcd bd7107bd4bf26219ba9852aa6c4c817ccde0191c in /home/pwaller/.local/src/v/cache/vcswork/e0603d6b678c67453e2d85ba58d5dd0030d8a3ed443832680252eb8f82272b52: exit status 1
vgo: stat github.com/prometheus/client_golang@52437c81da6b127a9925d17eb3a382a2e5fd395e: git fetch --depth=1 https://github.com/prometheus/client_golang 52437c81da6b127a9925d17eb3a382a2e5fd395e in /home/pwaller/.local/src/v/cache/vcswork/3785b9359e0a97a9dc849585696cf17189b86b713b7e59c2842fb3c44d3ff610: exit status 1
vgo: stat github.com/prometheus/client_model@fa8ad6fec33561be4280a8f0514318c79d7f6cb6: git fetch --depth=1 https://github.com/prometheus/client_model fa8ad6fec33561be4280a8f0514318c79d7f6cb6 in /home/pwaller/.local/src/v/cache/vcswork/2a98e665081184f4ca01f0af8738c882495d1fb131b7ed20ad844d3ba1bb6393: exit status 1
vgo: stat github.com/prometheus/common@ebdfc6da46522d58825777cf1f90490a5b1ef1d8: git fetch --depth=1 https://github.com/prometheus/common ebdfc6da46522d58825777cf1f90490a5b1ef1d8 in /home/pwaller/.local/src/v/cache/vcswork/78eeb4629eb558fa221be26a69bd8019d99c56c9d5e61a056019c7d4845bf714: exit status 1
vgo: stat github.com/prometheus/procfs@abf152e5f3e97f2fafac028d2cc06c1feb87ffa5: git fetch --depth=1 https://github.com/prometheus/procfs abf152e5f3e97f2fafac028d2cc06c1feb87ffa5 in /home/pwaller/.local/src/v/cache/vcswork/98c949bd9adf825ba7df08da3b4256328548aa917c603094bfc20a4042505d6d: exit status 1
vgo: stat github.com/docker/distribution@7230e9def796c63a4033211dc5107742d689fc1e: git fetch --depth=1 https://github.com/docker/distribution 7230e9def796c63a4033211dc5107742d689fc1e in /home/pwaller/.local/src/v/cache/vcswork/80b3acbee6ab6828cbdf1bb672722bc063a15e0f09bd3c6d10d932a2741bffd7: exit status 1
vgo: stat github.com/docker/docker@0fb0d67008157add34f1e11685e23a691db92644: git fetch --depth=1 https://github.com/docker/docker 0fb0d67008157add34f1e11685e23a691db92644 in /home/pwaller/.local/src/v/cache/vcswork/d0dd664b5d3114dfbbd886392787ccf2b8ea0c4e1f5371628ef4a2af00a5891e: exit status 1
vgo: stat github.com/docker/go-events@37d35add5005832485c0225ec870121b78fcff1c: git fetch --depth=1 https://github.com/docker/go-events 37d35add5005832485c0225ec870121b78fcff1c in /home/pwaller/.local/src/v/cache/vcswork/ceca82cd3bc327d14cd6a4b24d913c9734565bc15ec9d417711f0173f1d6dff0: exit status 1
vgo: stat github.com/docker/go-units@954fed01cc617c55d838fa2230073f2cb17386c8: git fetch --depth=1 https://github.com/docker/go-units 954fed01cc617c55d838fa2230073f2cb17386c8 in /home/pwaller/.local/src/v/cache/vcswork/9f2d1a527162210923891d0d8f65d59f937c6335cd9cad946f750ba191817910: exit status 1
vgo: stat github.com/docker/libkv@9fd56606e928ff1f309808f5d5a0b7a2ef73f9a8: git fetch --depth=1 https://github.com/docker/libkv 9fd56606e928ff1f309808f5d5a0b7a2ef73f9a8 in /home/pwaller/.local/src/v/cache/vcswork/8cf13d962f8420301cee1b5abcd347b4a05a814c506024123c43a12395242d7d: exit status 1
vgo: stat github.com/docker/libnetwork@3ab699ea36573d98f481d233c30c742ade737565: git fetch --depth=1 https://github.com/docker/libnetwork 3ab699ea36573d98f481d233c30c742ade737565 in /home/pwaller/.local/src/v/cache/vcswork/b336f1f9b64be3a88933393f9543b3e76e4a5bce0bb4b4ed42f9a075055e2a6e: exit status 1
vgo: stat github.com/opencontainers/runc@8e8d01d38d7b4fb0a35bf89b72bc3e18c98882d7: git fetch --depth=1 https://github.com/opencontainers/runc 8e8d01d38d7b4fb0a35bf89b72bc3e18c98882d7 in /home/pwaller/.local/src/v/cache/vcswork/2d235085b4e185f118a443ea632e2c4f8a6d96ef2b7f068acf187ba102f95dc2: exit status 1
vgo: stat github.com/davecgh/go-spew@5215b55f46b2b919f50a1df0eaa5886afe4e3b3d: git fetch --depth=1 https://github.com/davecgh/go-spew 5215b55f46b2b919f50a1df0eaa5886afe4e3b3d in /home/pwaller/.local/src/v/cache/vcswork/b9a4b9bbdb4a59723f2348415ad7ffda91568455a1cfd92e97976132bdfbaf57: exit status 1
vgo: stat github.com/Microsoft/go-winio@f778f05015353be65d242f3fedc18695756153bb: git fetch --depth=1 https://github.com/Microsoft/go-winio f778f05015353be65d242f3fedc18695756153bb in /home/pwaller/.local/src/v/cache/vcswork/af6525e610990b5a771d4f88755dc7ebf4d2198224a16772680f8fc46b15e9e0: exit status 1
vgo: stat github.com/Sirupsen/logrus@f76d643702a30fbffecdfe50831e11881c96ceb3: git fetch --depth=1 https://github.com/Sirupsen/logrus f76d643702a30fbffecdfe50831e11881c96ceb3 in /home/pwaller/.local/src/v/cache/vcswork/c2179b181e32ae463389bbbe6915c034015061c1d3c09d7159c54706feb29df9: exit status 1
vgo: stat github.com/beorn7/perks@4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9: git fetch --depth=1 https://github.com/beorn7/perks 4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9 in /home/pwaller/.local/src/v/cache/vcswork/1a437ed9cd95ef0b3850ad46f66cacade14a3fca4584dad97c8e5299ff7b4735: exit status 1
vgo: stat github.com/boltdb/bolt@e72f08ddb5a52992c0a44c7dda9316c7333938b2: git fetch --depth=1 https://github.com/boltdb/bolt e72f08ddb5a52992c0a44c7dda9316c7333938b2 in /home/pwaller/.local/src/v/cache/vcswork/e83c93b3c716fae8fb489742637e698e1bf98745aae3b5226ba355057bdf3717: exit status 1
vgo: stat github.com/cloudflare/cfssl@7fb22c8cba7ecaf98e4082d22d65800cf45e042a: git fetch --depth=1 https://github.com/cloudflare/cfssl 7fb22c8cba7ecaf98e4082d22d65800cf45e042a in /home/pwaller/.local/src/v/cache/vcswork/42a0dfebf1eff66b7c2d27c6fb5d7b2b0d6a8f266bbc7dd4fdbd5ac090201bf4: exit status 1
vgo: stat github.com/dustin/go-humanize@8929fe90cee4b2cb9deb468b51fb34eba64d1bf0: git fetch --depth=1 origin 8929fe90cee4b2cb9deb468b51fb34eba64d1bf0 in /home/pwaller/.local/src/v/cache/vcswork/adcb6c357515b30ad73495f8a066d4e01afbd25ee9ea2b8dd7ccc690a6c3e74a: exit status 1
vgo: stat github.com/golang/mock@bd3c8e81be01eef76d4b503f5e687d2d1354d2d9: git fetch --depth=1 https://github.com/golang/mock bd3c8e81be01eef76d4b503f5e687d2d1354d2d9 in /home/pwaller/.local/src/v/cache/vcswork/8fd2fb7f4befaedbf6e7d4c768d11fbe23bc1ce0d138e0689f91529c1898d9a1: exit status 1
vgo: stat github.com/google/certificate-transparency@0f6e3d1d1ba4d03fdaab7cd716f36255c2e48341: git fetch --depth=1 https://github.com/google/certificate-transparency 0f6e3d1d1ba4d03fdaab7cd716f36255c2e48341 in /home/pwaller/.local/src/v/cache/vcswork/91096961f6e56a56781818ef09bd9b5dabc8f04840ba433504b437194b01d65e: exit status 1
vgo: stat github.com/hashicorp/go-immutable-radix@8e8ed81f8f0bf1bdd829593fdd5c29922c1ea990: git fetch --depth=1 https://github.com/hashicorp/go-immutable-radix 8e8ed81f8f0bf1bdd829593fdd5c29922c1ea990 in /home/pwaller/.local/src/v/cache/vcswork/113258058091512d553dab74611129264ab512acc89e82d3e3cec1c07921f25e: exit status 1
vgo: stat github.com/hashicorp/go-memdb@608dda3b1410a73eaf3ac8b517c9ae7ebab6aa87: git fetch --depth=1 https://github.com/hashicorp/go-memdb 608dda3b1410a73eaf3ac8b517c9ae7ebab6aa87 in /home/pwaller/.local/src/v/cache/vcswork/242972787ae344da6341fa74d994440ce60c0ef2891b842d908c4c89ea9df05f: exit status 1
vgo: stat github.com/hashicorp/golang-lru@a0d98a5f288019575c6d1f4bb1573fef2d1fcdc4: git fetch --depth=1 https://github.com/hashicorp/golang-lru a0d98a5f288019575c6d1f4bb1573fef2d1fcdc4 in /home/pwaller/.local/src/v/cache/vcswork/1eaaa73a95627cc6e520c43f6cce496386f24ff1b7fa8af090724d7660d929e9: exit status 1
vgo: stat github.com/mreiferson/go-httpclient@63fe23f7434723dc904c901043af07931f293c47: git fetch --depth=1 https://github.com/mreiferson/go-httpclient 63fe23f7434723dc904c901043af07931f293c47 in /home/pwaller/.local/src/v/cache/vcswork/c2ec4676aae9fa411b819b9ed36e1e1bf7d5c88769b98537683db5205873675e: exit status 1
vgo: stat github.com/phayes/permbits@f7e3ac5e859d0b919c5068d581cc4c5d4f4f9bc5: git fetch --depth=1 origin f7e3ac5e859d0b919c5068d581cc4c5d4f4f9bc5 in /home/pwaller/.local/src/v/cache/vcswork/04fa29befc2b5d844150f0e4c4ca798cbb6f6a3ce963a0683f171a1f9cc73685: exit status 1
vgo: stat github.com/pivotal-golang/clock@3fd3c1944c59d9742e1cd333672181cd1a6f9fa0: git fetch --depth=1 https://github.com/pivotal-golang/clock 3fd3c1944c59d9742e1cd333672181cd1a6f9fa0 in /home/pwaller/.local/src/v/cache/vcswork/327a10928fdcae61e5fd5a109282c83f1a2bec4719a08ad511de884480edd964: exit status 1
vgo: stat github.com/rcrowley/go-metrics@51425a2415d21afadfd55cd93432c0bc69e9598d: git fetch --depth=1 origin 51425a2415d21afadfd55cd93432c0bc69e9598d in /home/pwaller/.local/src/v/cache/vcswork/1e49d5dd6fd83ee91bd92257edbbe27babfc2f75ebf19bb5d17f0e0029e46fd5: exit status 1
vgo: stat github.com/spf13/cobra@8e91712f174ced10270cf66615e0a9127e7c4de5: git fetch --depth=1 https://github.com/spf13/cobra 8e91712f174ced10270cf66615e0a9127e7c4de5 in /home/pwaller/.local/src/v/cache/vcswork/3acf82b7c983ee417907a837a4ec1200962dbab34a15385a11bc6f255dc04d6e: exit status 1
vgo: stat github.com/spf13/pflag@7f60f83a2c81bc3c3c0d5297f61ddfa68da9d3b7: git fetch --depth=1 https://github.com/spf13/pflag 7f60f83a2c81bc3c3c0d5297f61ddfa68da9d3b7 in /home/pwaller/.local/src/v/cache/vcswork/389cbbf79b0218a16e6f902e349b1cabca23e0203c06f228d24031e72b6cf480: exit status 1
vgo: finding golang.org/x/sys v0.0.0-20160222202601-5eaf0df67e70
vgo: finding golang.org/x/net v0.0.0-20160403195514-024ed629fd29
vgo: finding github.com/stretchr/testify v1.1.4
vgo: stat github.com/davecgh/go-spew@6d212800a42e8ab5c146b8ace3490ee17e5225f9: git fetch --depth=1 https://github.com/davecgh/go-spew 6d212800a42e8ab5c146b8ace3490ee17e5225f9 in /home/pwaller/.local/src/v/cache/vcswork/b9a4b9bbdb4a59723f2348415ad7ffda91568455a1cfd92e97976132bdfbaf57: exit status 1
vgo: stat github.com/pmezard/go-difflib@d8ed2627bdf02c080bf22230dbb337003b7aba2d: git fetch --depth=1 https://github.com/pmezard/go-difflib d8ed2627bdf02c080bf22230dbb337003b7aba2d in /home/pwaller/.local/src/v/cache/vcswork/9950c06efbb2d90e85a58f1fbd6f3eb2db497b7c539a93fb5555656c5aba3c13: exit status 1
vgo: stat github.com/stretchr/objx@cbeaeb16a013161a98496fad62933b1d21786672: git fetch --depth=1 origin cbeaeb16a013161a98496fad62933b1d21786672 in /home/pwaller/.local/src/v/cache/vcswork/b58cd1804573f08b6cfc86bbbad2960dd009cb14e98e8b74221958153f37a31b: exit status 1
vgo: finding github.com/pmezard/go-difflib v0.0.0-20160110105554-792786c7400a
vgo: finding github.com/pkg/errors v0.0.0-20160613021747-01fa4104b9c2
vgo: finding github.com/docker/go-connections v0.0.0-20160212231911-34b5052da6b1
vgo: finding github.com/coreos/pkg v0.0.0-20160727233714-3ac0863d7acf
vgo: finding github.com/coreos/go-systemd v0.0.0-20160826104600-43e4800a6165
vgo: finding github.com/docker/notary v0.4.2
vgo: stat github.com/Azure/go-ansiterm@388960b655244e76e24c75f48631564eaefade62: git fetch --depth=1 https://github.com/Azure/go-ansiterm 388960b655244e76e24c75f48631564eaefade62 in /home/pwaller/.local/src/v/cache/vcswork/da5145fda272732cd74527dacbe4967cc6a648dad514ccb83c288a28ea4c0671: exit status 1
vgo: stat github.com/Azure/go-ansiterm@388960b655244e76e24c75f48631564eaefade62: git fetch --depth=1 https://github.com/Azure/go-ansiterm 388960b655244e76e24c75f48631564eaefade62 in /home/pwaller/.local/src/v/cache/vcswork/da5145fda272732cd74527dacbe4967cc6a648dad514ccb83c288a28ea4c0671: exit status 1
vgo: stat github.com/BurntSushi/toml@bd2bdf7f18f849530ef7a1c29a4290217cab32a1: git fetch --depth=1 https://github.com/BurntSushi/toml bd2bdf7f18f849530ef7a1c29a4290217cab32a1 in /home/pwaller/.local/src/v/cache/vcswork/3c854d7dd8a65b0485436a210dc0dd7f98e1fe4cc8dc0bd3d77527fedda57561: exit status 1
vgo: stat github.com/BurntSushi/toml@bd2bdf7f18f849530ef7a1c29a4290217cab32a1: git fetch --depth=1 https://github.com/BurntSushi/toml bd2bdf7f18f849530ef7a1c29a4290217cab32a1 in /home/pwaller/.local/src/v/cache/vcswork/3c854d7dd8a65b0485436a210dc0dd7f98e1fe4cc8dc0bd3d77527fedda57561: exit status 1
vgo: stat github.com/BurntSushi/toml@bd2bdf7f18f849530ef7a1c29a4290217cab32a1: git fetch --depth=1 https://github.com/BurntSushi/toml bd2bdf7f18f849530ef7a1c29a4290217cab32a1 in /home/pwaller/.local/src/v/cache/vcswork/3c854d7dd8a65b0485436a210dc0dd7f98e1fe4cc8dc0bd3d77527fedda57561: exit status 1
vgo: stat github.com/BurntSushi/toml@bd2bdf7f18f849530ef7a1c29a4290217cab32a1: git fetch --depth=1 https://github.com/BurntSushi/toml bd2bdf7f18f849530ef7a1c29a4290217cab32a1 in /home/pwaller/.local/src/v/cache/vcswork/3c854d7dd8a65b0485436a210dc0dd7f98e1fe4cc8dc0bd3d77527fedda57561: exit status 1
vgo: stat github.com/Shopify/logrus-bugsnag@5a46080c635f13e8b60c24765c19d62e1ca8d0fb: git fetch --depth=1 origin 5a46080c635f13e8b60c24765c19d62e1ca8d0fb in /home/pwaller/.local/src/v/cache/vcswork/9281299b1fb7c23495802946626603799502d2d005854160cf1cca51c3988e75: exit status 1
vgo: stat github.com/Sirupsen/logrus@6d9ae300aaf85d6acd2e5424081c7fcddb21dab8: git fetch --depth=1 https://github.com/Sirupsen/logrus 6d9ae300aaf85d6acd2e5424081c7fcddb21dab8 in /home/pwaller/.local/src/v/cache/vcswork/c2179b181e32ae463389bbbe6915c034015061c1d3c09d7159c54706feb29df9: exit status 1
vgo: stat github.com/agl/ed25519@278e1ec8e8a6e017cd07577924d6766039146ced: git fetch --depth=1 https://github.com/agl/ed25519 278e1ec8e8a6e017cd07577924d6766039146ced in /home/pwaller/.local/src/v/cache/vcswork/0b82478dc41eb662b31914d0f711acabb12bf913cb746feb99c21463b792f537: exit status 1
vgo: stat github.com/agl/ed25519@278e1ec8e8a6e017cd07577924d6766039146ced: git fetch --depth=1 https://github.com/agl/ed25519 278e1ec8e8a6e017cd07577924d6766039146ced in /home/pwaller/.local/src/v/cache/vcswork/0b82478dc41eb662b31914d0f711acabb12bf913cb746feb99c21463b792f537: exit status 1
vgo: stat github.com/armon/consul-api@dcfedd50ed5334f96adee43fc88518a4f095e15c: git fetch --depth=1 origin dcfedd50ed5334f96adee43fc88518a4f095e15c in /home/pwaller/.local/src/v/cache/vcswork/f53dde7d7df3cddfeb4b159c4af03a63a930d1f7c2b4d4f21555d6b1fcd65df4: exit status 1
vgo: stat github.com/beorn7/perks@b965b613227fddccbfffe13eae360ed3fa822f8d: git fetch --depth=1 https://github.com/beorn7/perks b965b613227fddccbfffe13eae360ed3fa822f8d in /home/pwaller/.local/src/v/cache/vcswork/1a437ed9cd95ef0b3850ad46f66cacade14a3fca4584dad97c8e5299ff7b4735: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@13fd6b8acda029830ef9904df6b63be0a83369d0: git fetch --depth=1 origin 13fd6b8acda029830ef9904df6b63be0a83369d0 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@13fd6b8acda029830ef9904df6b63be0a83369d0: git fetch --depth=1 https://github.com/bugsnag/bugsnag-go 13fd6b8acda029830ef9904df6b63be0a83369d0 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@13fd6b8acda029830ef9904df6b63be0a83369d0: git fetch --depth=1 https://github.com/bugsnag/bugsnag-go 13fd6b8acda029830ef9904df6b63be0a83369d0 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@13fd6b8acda029830ef9904df6b63be0a83369d0: git fetch --depth=1 https://github.com/bugsnag/bugsnag-go 13fd6b8acda029830ef9904df6b63be0a83369d0 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@13fd6b8acda029830ef9904df6b63be0a83369d0: git fetch --depth=1 https://github.com/bugsnag/bugsnag-go 13fd6b8acda029830ef9904df6b63be0a83369d0 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@13fd6b8acda029830ef9904df6b63be0a83369d0: git fetch --depth=1 https://github.com/bugsnag/bugsnag-go 13fd6b8acda029830ef9904df6b63be0a83369d0 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@13fd6b8acda029830ef9904df6b63be0a83369d0: git fetch --depth=1 https://github.com/bugsnag/bugsnag-go 13fd6b8acda029830ef9904df6b63be0a83369d0 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@13fd6b8acda029830ef9904df6b63be0a83369d0: git fetch --depth=1 https://github.com/bugsnag/bugsnag-go 13fd6b8acda029830ef9904df6b63be0a83369d0 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/panicwrap@e2c28503fcd0675329da73bf48b33404db873782: git fetch --depth=1 origin e2c28503fcd0675329da73bf48b33404db873782 in /home/pwaller/.local/src/v/cache/vcswork/867a572df9f0f480ccde8c3f8a55799a224ea62a368baf143fdd67e1735235b2: exit status 1
vgo: stat github.com/cenkalti/backoff@4dc77674aceaabba2c7e3da25d4c823edfb73f99: git fetch --depth=1 origin 4dc77674aceaabba2c7e3da25d4c823edfb73f99 in /home/pwaller/.local/src/v/cache/vcswork/79e8d1963969a9df3425d256a195fff829ee975571acfe13306e2ad79ea04255: exit status 1
vgo: stat github.com/coreos/etcd@6acb3d67fbe131b3b2d5d010e00ec80182be4628: git fetch --depth=1 https://github.com/coreos/etcd 6acb3d67fbe131b3b2d5d010e00ec80182be4628 in /home/pwaller/.local/src/v/cache/vcswork/e0603d6b678c67453e2d85ba58d5dd0030d8a3ed443832680252eb8f82272b52: exit status 1
vgo: stat github.com/coreos/etcd@6acb3d67fbe131b3b2d5d010e00ec80182be4628: git fetch --depth=1 https://github.com/coreos/etcd 6acb3d67fbe131b3b2d5d010e00ec80182be4628 in /home/pwaller/.local/src/v/cache/vcswork/e0603d6b678c67453e2d85ba58d5dd0030d8a3ed443832680252eb8f82272b52: exit status 1
vgo: stat github.com/coreos/etcd@6acb3d67fbe131b3b2d5d010e00ec80182be4628: git fetch --depth=1 https://github.com/coreos/etcd 6acb3d67fbe131b3b2d5d010e00ec80182be4628 in /home/pwaller/.local/src/v/cache/vcswork/e0603d6b678c67453e2d85ba58d5dd0030d8a3ed443832680252eb8f82272b52: exit status 1
vgo: stat github.com/denisenkom/go-mssqldb@6e7f3d73dade2e5566f87d18c3a1d00d2ce33421: git fetch --depth=1 origin 6e7f3d73dade2e5566f87d18c3a1d00d2ce33421 in /home/pwaller/.local/src/v/cache/vcswork/d004f62f0bf917602489d74fdea89c7a3f58cffbf618ef1e4fc751f5d4836311: exit status 1
vgo: stat github.com/docker/docker@91853e44aeb20e55bcfcad5041c274783fdc06bc: git fetch --depth=1 https://github.com/docker/docker 91853e44aeb20e55bcfcad5041c274783fdc06bc in /home/pwaller/.local/src/v/cache/vcswork/d0dd664b5d3114dfbbd886392787ccf2b8ea0c4e1f5371628ef4a2af00a5891e: exit status 1
vgo: stat github.com/docker/docker@91853e44aeb20e55bcfcad5041c274783fdc06bc: git fetch --depth=1 https://github.com/docker/docker 91853e44aeb20e55bcfcad5041c274783fdc06bc in /home/pwaller/.local/src/v/cache/vcswork/d0dd664b5d3114dfbbd886392787ccf2b8ea0c4e1f5371628ef4a2af00a5891e: exit status 1
vgo: stat github.com/docker/docker@91853e44aeb20e55bcfcad5041c274783fdc06bc: git fetch --depth=1 https://github.com/docker/docker 91853e44aeb20e55bcfcad5041c274783fdc06bc in /home/pwaller/.local/src/v/cache/vcswork/d0dd664b5d3114dfbbd886392787ccf2b8ea0c4e1f5371628ef4a2af00a5891e: exit status 1
vgo: stat github.com/docker/go-connections@f549a9393d05688dff0992ef3efd8bbe6c628aeb: git fetch --depth=1 https://github.com/docker/go-connections f549a9393d05688dff0992ef3efd8bbe6c628aeb in /home/pwaller/.local/src/v/cache/vcswork/0380d1560cf386a65cff3e0bccbc134fcf7dc8f02de567ed81d3dd050231d328: exit status 1
vgo: stat github.com/docker/go-units@0bbddae09c5a5419a8c6dcdd7ff90da3d450393b: git fetch --depth=1 https://github.com/docker/go-units 0bbddae09c5a5419a8c6dcdd7ff90da3d450393b in /home/pwaller/.local/src/v/cache/vcswork/9f2d1a527162210923891d0d8f65d59f937c6335cd9cad946f750ba191817910: exit status 1
vgo: stat github.com/docker/libtrust@9cbd2a1374f46905c68a4eb3694a130610adc62a: git fetch --depth=1 https://github.com/docker/libtrust 9cbd2a1374f46905c68a4eb3694a130610adc62a in /home/pwaller/.local/src/v/cache/vcswork/da07fa0f90047c1906572bf6085c832cc0d3e6a5b251b2c1daacbbe31dba899f: exit status 1
vgo: stat github.com/docker/libtrust@9cbd2a1374f46905c68a4eb3694a130610adc62a: git fetch --depth=1 https://github.com/docker/libtrust 9cbd2a1374f46905c68a4eb3694a130610adc62a in /home/pwaller/.local/src/v/cache/vcswork/da07fa0f90047c1906572bf6085c832cc0d3e6a5b251b2c1daacbbe31dba899f: exit status 1
vgo: stat github.com/docker/libtrust@9cbd2a1374f46905c68a4eb3694a130610adc62a: git fetch --depth=1 https://github.com/docker/libtrust 9cbd2a1374f46905c68a4eb3694a130610adc62a in /home/pwaller/.local/src/v/cache/vcswork/da07fa0f90047c1906572bf6085c832cc0d3e6a5b251b2c1daacbbe31dba899f: exit status 1
vgo: stat github.com/docker/libtrust@9cbd2a1374f46905c68a4eb3694a130610adc62a: git fetch --depth=1 https://github.com/docker/libtrust 9cbd2a1374f46905c68a4eb3694a130610adc62a in /home/pwaller/.local/src/v/cache/vcswork/da07fa0f90047c1906572bf6085c832cc0d3e6a5b251b2c1daacbbe31dba899f: exit status 1
vgo: stat github.com/getsentry/raven-go@1cc47a9463b90f246a0503d4c2e9a55c9459ced3: git fetch --depth=1 origin 1cc47a9463b90f246a0503d4c2e9a55c9459ced3 in /home/pwaller/.local/src/v/cache/vcswork/e9dacfc63fb9364b4cdf606f260dee92f3758e3b0ce6122cf8c0176cb4da6243: exit status 1
vgo: stat github.com/go-sql-driver/mysql@0cc29e9fe8e25c2c58cf47bcab566e029bbaa88b: git fetch --depth=1 origin 0cc29e9fe8e25c2c58cf47bcab566e029bbaa88b in /home/pwaller/.local/src/v/cache/vcswork/f89e18d96ffb7c3b4b37fd44cdd3ae1a9ea27442eab27fe5ddf64b928fbfe286: exit status 1
vgo: stat github.com/golang/protobuf@c3cefd437628a0b7d31b34fe44b3a7a540e98527: git fetch --depth=1 https://github.com/golang/protobuf c3cefd437628a0b7d31b34fe44b3a7a540e98527 in /home/pwaller/.local/src/v/cache/vcswork/a20e27c072e7660590311c08d0311b908b66675cb634b0caf7cbb860e5c3c705: exit status 1
vgo: stat github.com/golang/protobuf@c3cefd437628a0b7d31b34fe44b3a7a540e98527: git fetch --depth=1 https://github.com/golang/protobuf c3cefd437628a0b7d31b34fe44b3a7a540e98527 in /home/pwaller/.local/src/v/cache/vcswork/a20e27c072e7660590311c08d0311b908b66675cb634b0caf7cbb860e5c3c705: exit status 1
vgo: stat github.com/google/gofuzz@bbcb9da2d746f8bdbd6a936686a0a6067ada0ec5: git fetch --depth=1 origin bbcb9da2d746f8bdbd6a936686a0a6067ada0ec5 in /home/pwaller/.local/src/v/cache/vcswork/494b36eea28d4b5317329e131e5193b575157a2ef53fa71b1b1175513e222398: exit status 1
vgo: stat github.com/gorilla/context@14f550f51af52180c2eefed15e5fd18d63c0a64a: git fetch --depth=1 https://github.com/gorilla/context 14f550f51af52180c2eefed15e5fd18d63c0a64a in /home/pwaller/.local/src/v/cache/vcswork/97ea9650f028793bd42d555dba7196a11b404ffbdbebffc0084e752976008464: exit status 1
vgo: stat github.com/gorilla/mux@e444e69cbd2e2e3e0749a2f3c717cec491552bbf: git fetch --depth=1 https://github.com/gorilla/mux e444e69cbd2e2e3e0749a2f3c717cec491552bbf in /home/pwaller/.local/src/v/cache/vcswork/6a3c85c1fa560af3c11df6bfabdc0a25c69c3127a5822875d586b38142c71189: exit status 1
vgo: stat github.com/jinzhu/gorm@82d726bbfd8cefbe2dcdc7f7f0484551c0d40433: git fetch --depth=1 origin 82d726bbfd8cefbe2dcdc7f7f0484551c0d40433 in /home/pwaller/.local/src/v/cache/vcswork/1915dd1d3c54bcaa4f52ef3dbfd9aa6f611de702acaf4d3373e322f5a49d9dc8: exit status 1
vgo: stat github.com/jinzhu/now@ce80572eb55aa0ac839330041ca9db1afa5f1f6c: git fetch --depth=1 origin ce80572eb55aa0ac839330041ca9db1afa5f1f6c in /home/pwaller/.local/src/v/cache/vcswork/e2f1237d1bd077d19ca49058f76399a08cf5dad5000ad96beabc0f79697bc081: exit status 1
vgo: stat github.com/juju/loggo@8477fc936adf0e382d680310047ca27e128a309a: git fetch --depth=1 origin 8477fc936adf0e382d680310047ca27e128a309a in /home/pwaller/.local/src/v/cache/vcswork/8fc4a12f0cf01496d94da41ce6967b85280c4e7d2ebf611ddde75ee7df358f2b: exit status 1
vgo: stat github.com/kr/pretty@bc9499caa0f45ee5edb2f0209fbd61fbf3d9018f: git fetch --depth=1 origin bc9499caa0f45ee5edb2f0209fbd61fbf3d9018f in /home/pwaller/.local/src/v/cache/vcswork/33119c1bd11a6422692760af70c33923f3c3f3c4373f6e1d9f93bed6ae1e1c89: exit status 1
vgo: stat github.com/kr/text@6807e777504f54ad073ecef66747de158294b639: git fetch --depth=1 origin 6807e777504f54ad073ecef66747de158294b639 in /home/pwaller/.local/src/v/cache/vcswork/3b687cf41b7eac3b8146bc2ccd6b61c15a059ee46970aaea9fd424243b742467: exit status 1
vgo: stat github.com/kr/text@6807e777504f54ad073ecef66747de158294b639: git fetch --depth=1 https://github.com/kr/text 6807e777504f54ad073ecef66747de158294b639 in /home/pwaller/.local/src/v/cache/vcswork/3b687cf41b7eac3b8146bc2ccd6b61c15a059ee46970aaea9fd424243b742467: exit status 1
vgo: stat github.com/kr/text@6807e777504f54ad073ecef66747de158294b639: git fetch --depth=1 https://github.com/kr/text 6807e777504f54ad073ecef66747de158294b639 in /home/pwaller/.local/src/v/cache/vcswork/3b687cf41b7eac3b8146bc2ccd6b61c15a059ee46970aaea9fd424243b742467: exit status 1
vgo: stat github.com/lib/pq@0dad96c0b94f8dee039aa40467f767467392a0af: git fetch --depth=1 origin 0dad96c0b94f8dee039aa40467f767467392a0af in /home/pwaller/.local/src/v/cache/vcswork/01b4997f7dd5821a38fc67ad1dbddb0839f3fe66c76533286b350cb97434f24c: exit status 1
vgo: stat github.com/lib/pq@0dad96c0b94f8dee039aa40467f767467392a0af: git fetch --depth=1 https://github.com/lib/pq 0dad96c0b94f8dee039aa40467f767467392a0af in /home/pwaller/.local/src/v/cache/vcswork/01b4997f7dd5821a38fc67ad1dbddb0839f3fe66c76533286b350cb97434f24c: exit status 1
vgo: stat github.com/lib/pq@0dad96c0b94f8dee039aa40467f767467392a0af: git fetch --depth=1 https://github.com/lib/pq 0dad96c0b94f8dee039aa40467f767467392a0af in /home/pwaller/.local/src/v/cache/vcswork/01b4997f7dd5821a38fc67ad1dbddb0839f3fe66c76533286b350cb97434f24c: exit status 1
vgo: stat github.com/matttproud/golang_protobuf_extensions@d0c3fe89de86839aecf2e0579c40ba3bb336a453: git fetch --depth=1 https://github.com/matttproud/golang_protobuf_extensions d0c3fe89de86839aecf2e0579c40ba3bb336a453 in /home/pwaller/.local/src/v/cache/vcswork/c9f9bbb8cc68e928d83601198e57e7c65ab57e6945016296f48e56d8e0f8d013: exit status 1
vgo: stat github.com/miekg/pkcs11@ba39b9c6300b7e0be41b115330145ef8afdff7d6: git fetch --depth=1 https://github.com/miekg/pkcs11 ba39b9c6300b7e0be41b115330145ef8afdff7d6 in /home/pwaller/.local/src/v/cache/vcswork/1d74d84db901415b73747d07b43f6cd36ac239739cfe9e183a4501b50e94c3ac: exit status 1
vgo: stat github.com/mitchellh/go-homedir@df55a15e5ce646808815381b3db47a8c66ea62f4: git fetch --depth=1 origin df55a15e5ce646808815381b3db47a8c66ea62f4 in /home/pwaller/.local/src/v/cache/vcswork/434f79a4b147edc55f3291424e4e021e69e3d8f7636dccb7d162f1b886e88dd5: exit status 1
vgo: stat github.com/mitchellh/mapstructure@2caf8efc93669b6c43e0441cdc6aed17546c96f3: git fetch --depth=1 https://github.com/mitchellh/mapstructure 2caf8efc93669b6c43e0441cdc6aed17546c96f3 in /home/pwaller/.local/src/v/cache/vcswork/cbe18bc04d35c650864e89ab9c27c6de5e24d51255fad0f64b974d13d4f47330: exit status 1
vgo: stat github.com/prometheus/client_golang@449ccefff16c8e2b7229f6be1921ba22f62461fe: git fetch --depth=1 https://github.com/prometheus/client_golang 449ccefff16c8e2b7229f6be1921ba22f62461fe in /home/pwaller/.local/src/v/cache/vcswork/3785b9359e0a97a9dc849585696cf17189b86b713b7e59c2842fb3c44d3ff610: exit status 1
vgo: stat github.com/prometheus/client_model@fa8ad6fec33561be4280a8f0514318c79d7f6cb6: git fetch --depth=1 https://github.com/prometheus/client_model fa8ad6fec33561be4280a8f0514318c79d7f6cb6 in /home/pwaller/.local/src/v/cache/vcswork/2a98e665081184f4ca01f0af8738c882495d1fb131b7ed20ad844d3ba1bb6393: exit status 1
vgo: stat github.com/prometheus/common@4fdc91a58c9d3696b982e8a680f4997403132d44: git fetch --depth=1 https://github.com/prometheus/common 4fdc91a58c9d3696b982e8a680f4997403132d44 in /home/pwaller/.local/src/v/cache/vcswork/78eeb4629eb558fa221be26a69bd8019d99c56c9d5e61a056019c7d4845bf714: exit status 1
vgo: stat github.com/prometheus/common@4fdc91a58c9d3696b982e8a680f4997403132d44: git fetch --depth=1 https://github.com/prometheus/common 4fdc91a58c9d3696b982e8a680f4997403132d44 in /home/pwaller/.local/src/v/cache/vcswork/78eeb4629eb558fa221be26a69bd8019d99c56c9d5e61a056019c7d4845bf714: exit status 1
vgo: stat github.com/prometheus/common@4fdc91a58c9d3696b982e8a680f4997403132d44: git fetch --depth=1 https://github.com/prometheus/common 4fdc91a58c9d3696b982e8a680f4997403132d44 in /home/pwaller/.local/src/v/cache/vcswork/78eeb4629eb558fa221be26a69bd8019d99c56c9d5e61a056019c7d4845bf714: exit status 1
vgo: stat github.com/prometheus/procfs@b1afdc266f54247f5dc725544f5d351a8661f502: git fetch --depth=1 https://github.com/prometheus/procfs b1afdc266f54247f5dc725544f5d351a8661f502 in /home/pwaller/.local/src/v/cache/vcswork/98c949bd9adf825ba7df08da3b4256328548aa917c603094bfc20a4042505d6d: exit status 1
vgo: stat github.com/revel/revel@a9a2ff45fae4330ef4116b257bcf9c82e53350c2: git fetch --depth=1 origin a9a2ff45fae4330ef4116b257bcf9c82e53350c2 in /home/pwaller/.local/src/v/cache/vcswork/a6201eff1865360d39d11d680a02d961fbf5553b1c3711131509787d49ffb09e: exit status 1
vgo: stat github.com/shurcooL/sanitized_anchor_name@244f5ac324cb97e1987ef901a0081a77bfd8e845: git fetch --depth=1 origin 244f5ac324cb97e1987ef901a0081a77bfd8e845 in /home/pwaller/.local/src/v/cache/vcswork/76c2951a1319a997cd8e8a30458383d1d32a6ff84eeca11610f09db33a517db9: exit status 1
vgo: stat github.com/spf13/cast@4d07383ffe94b5e5a6fa3af9211374a4507a0184: git fetch --depth=1 origin 4d07383ffe94b5e5a6fa3af9211374a4507a0184 in /home/pwaller/.local/src/v/cache/vcswork/5541c181005c14465a600ffd6cdf12b961c88117a036ba86fc7606ff1943a778: exit status 1
vgo: stat github.com/spf13/cobra@f368244301305f414206f889b1735a54cfc8bde8: git fetch --depth=1 https://github.com/spf13/cobra f368244301305f414206f889b1735a54cfc8bde8 in /home/pwaller/.local/src/v/cache/vcswork/3acf82b7c983ee417907a837a4ec1200962dbab34a15385a11bc6f255dc04d6e: exit status 1
vgo: stat github.com/spf13/jwalterweatherman@3d60171a64319ef63c78bd45bd60e6eab1e75f8b: git fetch --depth=1 origin 3d60171a64319ef63c78bd45bd60e6eab1e75f8b in /home/pwaller/.local/src/v/cache/vcswork/42f30feab47714451b083fa3fd4667b9f2db89ab045b1c405a65cfd2599d8689: exit status 1
vgo: stat github.com/spf13/pflag@cb88ea77998c3f024757528e3305022ab50b43be: git fetch --depth=1 https://github.com/spf13/pflag cb88ea77998c3f024757528e3305022ab50b43be in /home/pwaller/.local/src/v/cache/vcswork/389cbbf79b0218a16e6f902e349b1cabca23e0203c06f228d24031e72b6cf480: exit status 1
vgo: stat github.com/spf13/viper@be5ff3e4840cf692388bde7a057595a474ef379e: git fetch --depth=1 origin be5ff3e4840cf692388bde7a057595a474ef379e in /home/pwaller/.local/src/v/cache/vcswork/5fb3e42b1f6052aa08228f623b105973cd6121a1343230bfc0bf4b47f3528c7e: exit status 1
vgo: stat github.com/spf13/viper@be5ff3e4840cf692388bde7a057595a474ef379e: git fetch --depth=1 https://github.com/spf13/viper be5ff3e4840cf692388bde7a057595a474ef379e in /home/pwaller/.local/src/v/cache/vcswork/5fb3e42b1f6052aa08228f623b105973cd6121a1343230bfc0bf4b47f3528c7e: exit status 1
vgo: stat github.com/stevvooe/resumable@eb352b28d119500cb0382a8379f639c1c8d65831: git fetch --depth=1 origin eb352b28d119500cb0382a8379f639c1c8d65831 in /home/pwaller/.local/src/v/cache/vcswork/570205c992fe0bfb394f239e5eca70bc4fc956df5eef32859c74ea3113fa9944: exit status 1
vgo: stat github.com/stevvooe/resumable@eb352b28d119500cb0382a8379f639c1c8d65831: git fetch --depth=1 https://github.com/stevvooe/resumable eb352b28d119500cb0382a8379f639c1c8d65831 in /home/pwaller/.local/src/v/cache/vcswork/570205c992fe0bfb394f239e5eca70bc4fc956df5eef32859c74ea3113fa9944: exit status 1
vgo: stat github.com/stretchr/testify@089c7181b8c728499929ff09b62d3fdd8df8adff: git fetch --depth=1 https://github.com/stretchr/testify 089c7181b8c728499929ff09b62d3fdd8df8adff in /home/pwaller/.local/src/v/cache/vcswork/ed2f58bca3966d01dc4666baa48276a4fab360938a8d941050d58e371e2bba77: exit status 1
vgo: stat github.com/stretchr/testify@089c7181b8c728499929ff09b62d3fdd8df8adff: git fetch --depth=1 https://github.com/stretchr/testify 089c7181b8c728499929ff09b62d3fdd8df8adff in /home/pwaller/.local/src/v/cache/vcswork/ed2f58bca3966d01dc4666baa48276a4fab360938a8d941050d58e371e2bba77: exit status 1
vgo: stat github.com/stvp/go-udp-testing@06eb4f886d9f8242b0c176cf0d3ce5ec2cedda05: git fetch --depth=1 origin 06eb4f886d9f8242b0c176cf0d3ce5ec2cedda05 in /home/pwaller/.local/src/v/cache/vcswork/9ac5b016cf242df1d18fa169dd63d766c7c0b6882c2ae610804c884e9f415791: exit status 1
vgo: stat github.com/ugorji/go@c062049c1793b01a3cc3fe786108edabbaf7756b: git fetch --depth=1 https://github.com/ugorji/go c062049c1793b01a3cc3fe786108edabbaf7756b in /home/pwaller/.local/src/v/cache/vcswork/4c4ce012b2736486e99bf427f8d15912b07cd69667f52e5aac2af9a9f3c09a9e: exit status 1
vgo: stat github.com/xordataexchange/crypt@749e360c8f236773f28fc6d3ddfce4a470795227: git fetch --depth=1 origin 749e360c8f236773f28fc6d3ddfce4a470795227 in /home/pwaller/.local/src/v/cache/vcswork/aaedee3f2e0b2a1990367f896c2c42c789988795c7e7defd826b23e44a2b4d4e: exit status 1
vgo: stat github.com/xordataexchange/crypt@749e360c8f236773f28fc6d3ddfce4a470795227: git fetch --depth=1 https://github.com/xordataexchange/crypt 749e360c8f236773f28fc6d3ddfce4a470795227 in /home/pwaller/.local/src/v/cache/vcswork/aaedee3f2e0b2a1990367f896c2c42c789988795c7e7defd826b23e44a2b4d4e: exit status 1
vgo: stat github.com/xordataexchange/crypt@749e360c8f236773f28fc6d3ddfce4a470795227: git fetch --depth=1 https://github.com/xordataexchange/crypt 749e360c8f236773f28fc6d3ddfce4a470795227 in /home/pwaller/.local/src/v/cache/vcswork/aaedee3f2e0b2a1990367f896c2c42c789988795c7e7defd826b23e44a2b4d4e: exit status 1
vgo: stat github.com/xordataexchange/crypt@749e360c8f236773f28fc6d3ddfce4a470795227: git fetch --depth=1 https://github.com/xordataexchange/crypt 749e360c8f236773f28fc6d3ddfce4a470795227 in /home/pwaller/.local/src/v/cache/vcswork/aaedee3f2e0b2a1990367f896c2c42c789988795c7e7defd826b23e44a2b4d4e: exit status 1
vgo: stat github.com/xordataexchange/crypt@749e360c8f236773f28fc6d3ddfce4a470795227: git fetch --depth=1 https://github.com/xordataexchange/crypt 749e360c8f236773f28fc6d3ddfce4a470795227 in /home/pwaller/.local/src/v/cache/vcswork/aaedee3f2e0b2a1990367f896c2c42c789988795c7e7defd826b23e44a2b4d4e: exit status 1
vgo: stat google.golang.org/appengine@41265fb44deca5c3b05a946d5db1f54ae54fe67e: git fetch --depth=1 origin 41265fb44deca5c3b05a946d5db1f54ae54fe67e in /home/pwaller/.local/src/v/cache/vcswork/606c3f865bae05dafcee0ab4f2fcffe95f20e2ab6a3430cd814a17d4baa50fe2: exit status 1
vgo: lookup google.golang.org/appengine/internal: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/app_identity: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/base: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/datastore: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/log: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/modules: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/remote_api: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/cloud/compute/metadata: module root is "google.golang.org/cloud"
vgo: lookup google.golang.org/cloud/internal: module root is "google.golang.org/cloud"
vgo: stat google.golang.org/grpc@452f01f3ae159dcec83bbf0c6b0d96860e07b5e6: git fetch --depth=1 https://github.com/grpc/grpc-go 452f01f3ae159dcec83bbf0c6b0d96860e07b5e6 in /home/pwaller/.local/src/v/cache/vcswork/53ab5f2f034ba42de32f909aa45670cf730847987f38664c4052b329152ad727: exit status 1
vgo: lookup google.golang.org/grpc/benchmark: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/benchmark/client: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/benchmark/grpc_testing: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/benchmark/server: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/benchmark/stats: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/codes: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/credentials: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/credentials/oauth: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/examples/helloworld/greeter_client: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/examples/helloworld/greeter_server: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/examples/helloworld/helloworld: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/examples/route_guide/client: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/examples/route_guide/routeguide: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/examples/route_guide/server: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/grpclog: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/grpclog/glogger: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/health: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/interop/client: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/interop/grpc_testing: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/interop/server: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/metadata: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/naming: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/test/codec_perf: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/test/grpc_testing: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/transport: module root is "google.golang.org/grpc"
vgo: lookup gopkg.in/dancannon/gorethink.v2/encoding: invalid module path "gopkg.in/dancannon/gorethink.v2/encoding"
vgo: lookup gopkg.in/dancannon/gorethink.v2/ql2: invalid module path "gopkg.in/dancannon/gorethink.v2/ql2"
vgo: lookup gopkg.in/dancannon/gorethink.v2/types: invalid module path "gopkg.in/dancannon/gorethink.v2/types"
vgo: stat github.com/golang/protobuf@c3cefd437628a0b7d31b34fe44b3a7a540e98527: git fetch --depth=1 https://github.com/golang/protobuf c3cefd437628a0b7d31b34fe44b3a7a540e98527 in /home/pwaller/.local/src/v/cache/vcswork/a20e27c072e7660590311c08d0311b908b66675cb634b0caf7cbb860e5c3c705: exit status 1
vgo: lookup google.golang.org/grpc/internal: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/peer: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/health/grpc_health_v1: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/interop: module root is "google.golang.org/grpc"
vgo: finding gopkg.in/yaml.v2 v2.0.0-20150116202057-bef53efd0c76
vgo: finding gopkg.in/fsnotify.v1 v1.0.0-20160303034835-875cf421b32f
vgo: finding gopkg.in/fatih/pool.v2 v2.0.0-20150325163252-cba550ebf9bc
vgo: finding gopkg.in/dancannon/gorethink.v2 v2.0.0-20160418200057-3742792da4bc
vgo: finding gopkg.in/check.v1 v1.0.0-20160105164936-4f90aeace3a2
vgo: finding golang.org/x/sys v0.0.0-20160113011410-442cd600860c
vgo: finding golang.org/x/oauth2 v0.0.0-20160323034610-93758b5cba8c
vgo: finding golang.org/x/net v0.0.0-20160726221601-6a513affb38d
vgo: finding golang.org/x/crypto v0.0.0-20160518162255-5bcd134fee4d
vgo: finding github.com/tobi/airbrake-go v0.0.0-20151005181455-a3cdd910a3ff
vgo: finding github.com/russross/blackfriday v0.0.0-20150720194836-8cec3a854e68
vgo: finding github.com/robfig/pathtree v0.0.0-20140121041023-41257a1839e9
vgo: finding github.com/robfig/config v0.0.0-20141207224736-0f78529c8c7e
vgo: finding github.com/mattn/go-sqlite3 v0.0.0-20150629235728-b4142c444a89
vgo: finding github.com/magiconair/properties v0.0.0-20150601221201-624009598839
vgo: finding github.com/kr/pty v0.0.0-20151007230424-f7ee69f31298
vgo: finding github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed
vgo: finding github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
vgo: finding github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5
vgo: finding github.com/dvsekhvalnov/jose2go v0.0.0-20150816204322-6387d3c1f5ab
vgo: finding github.com/docker/go v0.0.0-20160303222718-d30aec9fd63c
vgo: finding github.com/docker/distribution v0.0.0-20160826182037-12acdf0a6c1e
vgo: stat github.com/Azure/azure-sdk-for-go@95361a2573b1fa92a00c5fc2707a80308483c6f9: git fetch --depth=1 origin 95361a2573b1fa92a00c5fc2707a80308483c6f9 in /home/pwaller/.local/src/v/cache/vcswork/ddc1fce0457234351c5843ac39f6d92a222dfea36f9945c4065f7aef74acf393: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/aws/aws-sdk-go@49c3892b61af1d4996292a3025f36e4dfa25eaee: git fetch --depth=1 https://github.com/aws/aws-sdk-go 49c3892b61af1d4996292a3025f36e4dfa25eaee in /home/pwaller/.local/src/v/cache/vcswork/cb1953cbdfd14fc2ffce4dfd06487e8d5a0c96da00d00bdef65874ff644eaa26: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@b1d153021fcd90ca3f080db36bec96dc690fb274: git fetch --depth=1 https://github.com/bugsnag/bugsnag-go b1d153021fcd90ca3f080db36bec96dc690fb274 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@b1d153021fcd90ca3f080db36bec96dc690fb274: git fetch --depth=1 https://github.com/bugsnag/bugsnag-go b1d153021fcd90ca3f080db36bec96dc690fb274 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/panicwrap@e2c28503fcd0675329da73bf48b33404db873782: git fetch --depth=1 https://github.com/bugsnag/panicwrap e2c28503fcd0675329da73bf48b33404db873782 in /home/pwaller/.local/src/v/cache/vcswork/867a572df9f0f480ccde8c3f8a55799a224ea62a368baf143fdd67e1735235b2: exit status 1
vgo: stat github.com/denverdino/aliyungo@6ffb587da9da6d029d0ce517b85fecc82172d502: git fetch --depth=1 origin 6ffb587da9da6d029d0ce517b85fecc82172d502 in /home/pwaller/.local/src/v/cache/vcswork/87de305cbe98f03fac8b44c5dc2ddcf88eb6b7c709c3a8d71cef35da2b5e5b8c: exit status 1
vgo: stat github.com/denverdino/aliyungo@6ffb587da9da6d029d0ce517b85fecc82172d502: git fetch --depth=1 https://github.com/denverdino/aliyungo 6ffb587da9da6d029d0ce517b85fecc82172d502 in /home/pwaller/.local/src/v/cache/vcswork/87de305cbe98f03fac8b44c5dc2ddcf88eb6b7c709c3a8d71cef35da2b5e5b8c: exit status 1
vgo: stat github.com/denverdino/aliyungo@6ffb587da9da6d029d0ce517b85fecc82172d502: git fetch --depth=1 https://github.com/denverdino/aliyungo 6ffb587da9da6d029d0ce517b85fecc82172d502 in /home/pwaller/.local/src/v/cache/vcswork/87de305cbe98f03fac8b44c5dc2ddcf88eb6b7c709c3a8d71cef35da2b5e5b8c: exit status 1
vgo: stat github.com/docker/libtrust@fa567046d9b14f6aa788882a950d69651d230b21: git fetch --depth=1 https://github.com/docker/libtrust fa567046d9b14f6aa788882a950d69651d230b21 in /home/pwaller/.local/src/v/cache/vcswork/da07fa0f90047c1906572bf6085c832cc0d3e6a5b251b2c1daacbbe31dba899f: exit status 1
vgo: stat github.com/garyburd/redigo@535138d7bcd717d6531c701ef5933d98b1866257: git fetch --depth=1 origin 535138d7bcd717d6531c701ef5933d98b1866257 in /home/pwaller/.local/src/v/cache/vcswork/ac29ab50a4fbaf6da97387824411aaa95a5364a02f9398964da7b5ad27006c30: exit status 1
vgo: stat github.com/garyburd/redigo@535138d7bcd717d6531c701ef5933d98b1866257: git fetch --depth=1 https://github.com/garyburd/redigo 535138d7bcd717d6531c701ef5933d98b1866257 in /home/pwaller/.local/src/v/cache/vcswork/ac29ab50a4fbaf6da97387824411aaa95a5364a02f9398964da7b5ad27006c30: exit status 1
vgo: stat github.com/golang/protobuf@8d92cf5fc15a4382f8964b08e1f42a75c0591aa3: git fetch --depth=1 https://github.com/golang/protobuf 8d92cf5fc15a4382f8964b08e1f42a75c0591aa3 in /home/pwaller/.local/src/v/cache/vcswork/a20e27c072e7660590311c08d0311b908b66675cb634b0caf7cbb860e5c3c705: exit status 1
vgo: stat github.com/gorilla/context@14f550f51af52180c2eefed15e5fd18d63c0a64a: git fetch --depth=1 https://github.com/gorilla/context 14f550f51af52180c2eefed15e5fd18d63c0a64a in /home/pwaller/.local/src/v/cache/vcswork/97ea9650f028793bd42d555dba7196a11b404ffbdbebffc0084e752976008464: exit status 1
vgo: stat github.com/gorilla/handlers@60c7bfde3e33c201519a200a4507a158cc03a17b: git fetch --depth=1 origin 60c7bfde3e33c201519a200a4507a158cc03a17b in /home/pwaller/.local/src/v/cache/vcswork/9afcc1b94a52c44fa45724d4dde25bb668b98b272dd07a4f4ce110f737f1a9c6: exit status 1
vgo: stat github.com/gorilla/mux@e444e69cbd2e2e3e0749a2f3c717cec491552bbf: git fetch --depth=1 https://github.com/gorilla/mux e444e69cbd2e2e3e0749a2f3c717cec491552bbf in /home/pwaller/.local/src/v/cache/vcswork/6a3c85c1fa560af3c11df6bfabdc0a25c69c3127a5822875d586b38142c71189: exit status 1
vgo: stat github.com/jmespath/go-jmespath@0b12d6b521d83fc7f755e7cfc1b1fbdd35a01a74: git fetch --depth=1 https://github.com/jmespath/go-jmespath 0b12d6b521d83fc7f755e7cfc1b1fbdd35a01a74 in /home/pwaller/.local/src/v/cache/vcswork/7b1106ecb177564b0bc9784f963c6c785e31d09dcd9f08114684d32af620443f: exit status 1
vgo: stat github.com/mitchellh/mapstructure@482a9fd5fa83e8c4e7817413b80f3eb8feec03ef: git fetch --depth=1 https://github.com/mitchellh/mapstructure 482a9fd5fa83e8c4e7817413b80f3eb8feec03ef in /home/pwaller/.local/src/v/cache/vcswork/cbe18bc04d35c650864e89ab9c27c6de5e24d51255fad0f64b974d13d4f47330: exit status 1
vgo: stat github.com/ncw/swift@ce444d6d47c51d4dda9202cd38f5094dd8e27e86: git fetch --depth=1 origin ce444d6d47c51d4dda9202cd38f5094dd8e27e86 in /home/pwaller/.local/src/v/cache/vcswork/6e867106694e212de97e2f2769afc1c7c0d23041c3a8135b37ec70c27b52943b: exit status 1
vgo: stat github.com/ncw/swift@ce444d6d47c51d4dda9202cd38f5094dd8e27e86: git fetch --depth=1 https://github.com/ncw/swift ce444d6d47c51d4dda9202cd38f5094dd8e27e86 in /home/pwaller/.local/src/v/cache/vcswork/6e867106694e212de97e2f2769afc1c7c0d23041c3a8135b37ec70c27b52943b: exit status 1
vgo: stat github.com/spf13/cobra@312092086bed4968099259622145a0c9ae280064: git fetch --depth=1 https://github.com/spf13/cobra 312092086bed4968099259622145a0c9ae280064 in /home/pwaller/.local/src/v/cache/vcswork/3acf82b7c983ee417907a837a4ec1200962dbab34a15385a11bc6f255dc04d6e: exit status 1
vgo: stat github.com/spf13/pflag@5644820622454e71517561946e3d94b9f9db6842: git fetch --depth=1 https://github.com/spf13/pflag 5644820622454e71517561946e3d94b9f9db6842 in /home/pwaller/.local/src/v/cache/vcswork/389cbbf79b0218a16e6f902e349b1cabca23e0203c06f228d24031e72b6cf480: exit status 1
vgo: stat github.com/stevvooe/resumable@51ad44105773cafcbe91927f70ac68e1bf78f8b4: git fetch --depth=1 https://github.com/stevvooe/resumable 51ad44105773cafcbe91927f70ac68e1bf78f8b4 in /home/pwaller/.local/src/v/cache/vcswork/570205c992fe0bfb394f239e5eca70bc4fc956df5eef32859c74ea3113fa9944: exit status 1
vgo: stat github.com/stevvooe/resumable@51ad44105773cafcbe91927f70ac68e1bf78f8b4: git fetch --depth=1 https://github.com/stevvooe/resumable 51ad44105773cafcbe91927f70ac68e1bf78f8b4 in /home/pwaller/.local/src/v/cache/vcswork/570205c992fe0bfb394f239e5eca70bc4fc956df5eef32859c74ea3113fa9944: exit status 1
vgo: stat github.com/stevvooe/resumable@51ad44105773cafcbe91927f70ac68e1bf78f8b4: git fetch --depth=1 https://github.com/stevvooe/resumable 51ad44105773cafcbe91927f70ac68e1bf78f8b4 in /home/pwaller/.local/src/v/cache/vcswork/570205c992fe0bfb394f239e5eca70bc4fc956df5eef32859c74ea3113fa9944: exit status 1
vgo: stat github.com/yvasiyarov/go-metrics@57bccd1ccd43f94bb17fdd8bf3007059b802f85e: git fetch --depth=1 origin 57bccd1ccd43f94bb17fdd8bf3007059b802f85e in /home/pwaller/.local/src/v/cache/vcswork/abd2823bce0a6598ea32537fabfd9cd2d0f389cfd023f319df5e05d8977b2fb6: exit status 1
vgo: stat github.com/yvasiyarov/gorelic@a9bba5b9ab508a086f9a12b8c51fab68478e2128: git fetch --depth=1 origin a9bba5b9ab508a086f9a12b8c51fab68478e2128 in /home/pwaller/.local/src/v/cache/vcswork/3a3ef5c2da0642f123b8cb28aa914d49f2fa618e45adddfd40b821560787999f: exit status 1
vgo: stat github.com/yvasiyarov/newrelic_platform_go@b21fdbd4370f3717f3bbd2bf41c223bc273068e6: git fetch --depth=1 origin b21fdbd4370f3717f3bbd2bf41c223bc273068e6 in /home/pwaller/.local/src/v/cache/vcswork/b244b25b60c150ec4f0f7d9f8654f09a9abc8fa82d57b52ba48ed048cb02fd3d: exit status 1
vgo: lookup google.golang.org/api/gensupport: module root is "google.golang.org/api"
vgo: lookup google.golang.org/api/googleapi: module root is "google.golang.org/api"
vgo: lookup google.golang.org/api/googleapi/internal/uritemplates: module root is "google.golang.org/api"
vgo: lookup google.golang.org/api/storage/v1: module root is "google.golang.org/api"
vgo: stat google.golang.org/appengine@12d5545dc1cfa6047a286d5e853841b6471f4c19: git fetch --depth=1 https://github.com/golang/appengine 12d5545dc1cfa6047a286d5e853841b6471f4c19 in /home/pwaller/.local/src/v/cache/vcswork/606c3f865bae05dafcee0ab4f2fcffe95f20e2ab6a3430cd814a17d4baa50fe2: exit status 1
vgo: lookup google.golang.org/appengine/internal: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/app_identity: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/base: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/datastore: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/log: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/modules: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/appengine/internal/remote_api: module root is "google.golang.org/appengine"
vgo: lookup google.golang.org/cloud/compute/metadata: module root is "google.golang.org/cloud"
vgo: lookup google.golang.org/cloud/internal: module root is "google.golang.org/cloud"
vgo: lookup google.golang.org/cloud/internal/opts: module root is "google.golang.org/cloud"
vgo: lookup google.golang.org/cloud/storage: module root is "google.golang.org/cloud"
vgo: stat google.golang.org/grpc@d3ddb4469d5a1b949fc7a7da7c1d6a0d1b6de994: git fetch --depth=1 https://github.com/grpc/grpc-go d3ddb4469d5a1b949fc7a7da7c1d6a0d1b6de994 in /home/pwaller/.local/src/v/cache/vcswork/53ab5f2f034ba42de32f909aa45670cf730847987f38664c4052b329152ad727: exit status 1
vgo: lookup google.golang.org/grpc/codes: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/credentials: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/grpclog: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/internal: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/metadata: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/naming: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/peer: module root is "google.golang.org/grpc"
vgo: lookup google.golang.org/grpc/transport: module root is "google.golang.org/grpc"
vgo: stat rsc.io/letsencrypt@a019c9e6fce0c7132679dea13bd8df7c86ffe26c: git fetch --depth=1 origin a019c9e6fce0c7132679dea13bd8df7c86ffe26c in /home/pwaller/.local/src/v/cache/vcswork/092b467b1b246260ef830bdfbfaefa977c63414385ce3b90695546a830e4e77b: exit status 1
vgo: finding gopkg.in/check.v1 v1.0.0-20141024133853-64131543e789
vgo: finding google.golang.org/cloud v0.0.0-20151119220103-975617b05ea8
vgo: finding golang.org/x/oauth2 v0.0.0-20160304213135-045497edb623
vgo: finding golang.org/x/net v0.0.0-20160322021652-4876518f9e71
vgo: finding golang.org/x/crypto v0.0.0-20150531185727-c10c31b5e94b
vgo: finding github.com/go-ini/ini v0.0.0-20160106105616-afbd495e5aae
vgo: finding github.com/docker/goamz v0.0.0-20160206023946-f0a21f5b2e12
vgo: stat github.com/cbroglie/mapstructure@25325b46b67d1c3eb7d58bad37d34d89a31cf9ec: git fetch --depth=1 origin 25325b46b67d1c3eb7d58bad37d34d89a31cf9ec in /home/pwaller/.local/src/v/cache/vcswork/6e007e5df6472f708617d54e302c969b268dd0cd97d66f6fe0f43c68e997d965: exit status 1
vgo: stat github.com/feyeleanor/raw@724aedf6e1a5d8971aafec384b6bde3d5608fba4: git fetch --depth=1 origin 724aedf6e1a5d8971aafec384b6bde3d5608fba4 in /home/pwaller/.local/src/v/cache/vcswork/41c2c0be69ac2d89233fa2bd17c0e0c34259427289da5001a56e896e50883b28: exit status 1
vgo: stat github.com/feyeleanor/slices@bb44bb2e4817fe71ba7082d351fd582e7d40e3ea: git fetch --depth=1 origin bb44bb2e4817fe71ba7082d351fd582e7d40e3ea in /home/pwaller/.local/src/v/cache/vcswork/6ff21d8d90fbd794b1bb3166ea19b671e1abcb7c53d8444e5172427cf79521bd: exit status 1
vgo: finding github.com/feyeleanor/sets v0.0.0-20130227004140-6c54cb57ea40
vgo: finding github.com/Sirupsen/logrus v0.0.0-20150409230825-55eb11d21d2a
vgo: finding github.com/cpuguy83/go-md2man v0.0.0-20150803153522-71acacd42f85
vgo: finding github.com/coreos/go-etcd v0.0.0-20151026160318-003851be7bb0
vgo: finding github.com/bugsnag/osext v0.0.0-20130617224835-0dd3f918b21b
vgo: finding github.com/bitly/go-simplejson v0.0.0-20150915165335-aabad6e81978
vgo: finding github.com/agtorre/gocolorize v0.0.0-20140921225620-f42b554bf7f0
vgo: finding github.com/docker/go-connections v0.0.0-20170204000113-ecb4cb2dd420
vgo: finding github.com/docker/distribution v0.0.0-20161216195125-28602af35ace
vgo: stat github.com/Azure/azure-sdk-for-go@0b5fe2abe0271ba07049eacaa65922d67c319543: git fetch --depth=1 https://github.com/Azure/azure-sdk-for-go 0b5fe2abe0271ba07049eacaa65922d67c319543 in /home/pwaller/.local/src/v/cache/vcswork/ddc1fce0457234351c5843ac39f6d92a222dfea36f9945c4065f7aef74acf393: exit status 1
vgo: stat github.com/bugsnag/bugsnag-go@b1d153021fcd90ca3f080db36bec96dc690fb274: git fetch --depth=1 https://github.com/bugsnag/bugsnag-go b1d153021fcd90ca3f080db36bec96dc690fb274 in /home/pwaller/.local/src/v/cache/vcswork/d23dd1f42742621bc6774d4e70035073c305013899beba81399197ac0f458606: exit status 1
vgo: stat github.com/bugsnag/panicwrap@e2c28503fcd0675329da73bf48b33404db873782: git fetch --depth=1 https://github.com/bugsnag/panicwrap e2c28503fcd0675329da73bf48b33404db873782 in /home/pwaller/.local/src/v/cache/vcswork/867a572df9f0f480ccde8c3f8a55799a224ea62a368baf143fdd67e1735235b2: exit status 1
vgo: stat github.com/denverdino/aliyungo@afedced274aa9a7fcdd47ac97018f0f8db4e5de2: git fetch --depth=1 https://github.com/denverdino/aliyungo afedced274aa9a7fcdd47ac97018f0f8db4e5de2 in /home/pwaller/.local/src/v/cache/vcswork/87de305cbe98f03fac8b44c5dc2ddcf88eb6b7c709c3a8d71cef35da2b5e5b8c: exit status 1
vgo: stat github.com/docker/libtrust@fa567046d9b14f6aa788882a950d69651d230b21: git fetch --depth=1 https://github.com/docker/libtrust fa567046d9b14f6aa788882a950d69651d230b21 in /home/pwaller/.local/src/v/cache/vcswork/da07fa0f90047c1906572bf6085c832cc0d3e6a5b251b2c1daacbbe31dba899f: exit status 1
vgo: stat github.com/garyburd/redigo@535138d7bcd717d6531c701ef5933d98b1866257: git fetch --depth=1 https://github.com/garyburd/redigo 535138d7bcd717d6531c701ef5933d98b1866257 in /home/pwaller/.local/src/v/cache/vcswork/ac29ab50a4fbaf6da97387824411aaa95a5364a02f9398964da7b5ad27006c30: exit status 1
vgo: stat github.com/go-ini/ini@2ba15ac2dc9cdf88c110ec2dc0ced7fa45f5678c: git fetch --depth=1 https://github.com/go-ini/ini 2ba15ac2dc9cdf88c110ec2dc0ced7fa45f5678c in /home/pwaller/.local/src/v/cache/vcswork/f63afdd23e3f4a0ba51aa6e624a654abefeae191beca1b19ebd2c902c09ef9d0: exit status 1
vgo: stat github.com/golang/protobuf@8d92cf5fc15a4382f8964b08e1f42a75c0591aa3: git fetch --depth=1 https://github.com/golang/protobuf 8d92cf5fc15a4382f8964b08e1f42a75c0591aa3 in /home/pwaller/.local/src/v/cache/vcswork/a20e27c072e7660590311c08d0311b908b66675cb634b0caf7cbb860e5c3c705: exit status 1
vgo: stat github.com/gorilla/context@14f550f51af52180c2eefed15e5fd18d63c0a64a: git fetch --depth=1 https://github.com/gorilla/context 14f550f51af52180c2eefed15e5fd18d63c0a64a in /home/pwaller/.local/src/v/cache/vcswork/97ea9650f028793bd42d555dba7196a11b404ffbdbebffc0084e752976008464: exit status 1
vgo: stat github.com/gorilla/handlers@60c7bfde3e33c201519a200a4507a158cc03a17b: git fetch --depth=1 https://github.com/gorilla/handlers 60c7bfde3e33c201519a200a4507a158cc03a17b in /home/pwaller/.local/src/v/cache/vcswork/9afcc1b94a52c44fa45724d4dde25bb668b98b272dd07a4f4ce110f737f1a9c6: exit status 1
vgo: stat github.com/gorilla/mux@e444e69cbd2e2e3e0749a2f3c717cec491552bbf: git fetch --depth=1 https://github.com/gorilla/mux e444e69cbd2e2e3e0749a2f3c717cec491552bbf in /home/pwaller/.local/src/v/cache/vcswork/6a3c85c1fa560af3c11df6bfabdc0a25c69c3127a5822875d586b38142c71189: exit status 1
vgo: stat github.com/jmespath/go-jmespath@bd40a432e4c76585ef6b72d3fd96fb9b6dc7b68d: git fetch --depth=1 https://github.com/jmespath/go-jmespath bd40a432e4c76585ef6b72d3fd96fb9b6dc7b68d in /home/pwaller/.local/src/v/cache/vcswork/7b1106ecb177564b0bc9784f963c6c785e31d09dcd9f08114684d32af620443f: exit status 1
vgo: stat github.com/miekg/dns@271c58e0c14f552178ea321a545ff9af38930f39: git fetch --depth=1 https://github.com/miekg/dns 271c58e0c14f552178ea321a545ff9af38930f39 in /home/pwaller/.local/src/v/cache/vcswork/8398ef9e2385ed1171ebb49f31e4200d019135b91a0387e3cc8af4c7405500d0: exit status 1
vgo: stat github.com/mitchellh/mapstructure@482a9fd5fa83e8c4e7817413b80f3eb8feec03ef: git fetch --depth=1 https://github.com/mitchellh/mapstructure 482a9fd5fa83e8c4e7817413b80f3eb8feec03ef in /home/pwaller/.local/src/v/cache/vcswork/cbe18bc04d35c650864e89ab9c27c6de5e24d51255fad0f64b974d13d4f47330: exit status 1
vgo: stat github.com/ncw/swift@b964f2ca856aac39885e258ad25aec08d5f64ee6: git fetch --depth=1 https://github.com/ncw/swift b964f2ca856aac39885e258ad25aec08d5f64ee6 in /home/pwaller/.local/src/v/cache/vcswork/6e867106694e212de97e2f2769afc1c7c0d23041c3a8135b37ec70c27b52943b: exit status 1
vgo: stat github.com/spf13/cobra@312092086bed4968099259622145a0c9ae280064: git fetch --depth=1 https://github.com/spf13/cobra 312092086bed4968099259622145a0c9ae280064 in /home/pwaller/.local/src/v/cache/vcswork/3acf82b7c983ee417907a837a4ec1200962dbab34a15385a11bc6f255dc04d6e: exit status 1
vgo: stat github.com/spf13/pflag@5644820622454e71517561946e3d94b9f9db6842: git fetch --depth=1 https://github.com/spf13/pflag 5644820622454e71517561946e3d94b9f9db6842 in /home/pwaller/.local/src/v/cache/vcswork/389cbbf79b0218a16e6f902e349b1cabca23e0203c06f228d24031e72b6cf480: exit status 1
vgo: stat github.com/stevvooe/resumable@51ad44105773cafcbe91927f70ac68e1bf78f8b4: git fetch --depth=1 https://github.com/stevvooe/resumable 51ad44105773cafcbe91927f70ac68e1bf78f8b4 in /home/pwaller/.local/src/v/cache/vcswork/570205c992fe0bfb394f239e5eca70bc4fc956df5eef32859c74ea3113fa9944: exit status 1
vgo: stat github.com/xenolf/lego@a9d8cec0e6563575e5868a005359ac97911b5985: git fetch --depth=1 origin a9d8cec0e6563575e5868a005359ac97911b5985 in /home/pwaller/.local/src/v/cache/vcswork/53cf8baf262f6c74b7e374be8a6879ce6afd71b62d8b081643918f9a8ef33ee1: exit status 1
vgo: stat github.com/yvasiyarov/go-metrics@57bccd1ccd43f94bb17fdd8bf3007059b802f85e: git fetch --depth=1 https://github.com/yvasiyarov/go-metrics 57bccd1ccd43f94bb17fdd8bf3007059b802f85e in /home/pwaller/.local/src/v/cache/vcswork/abd2823bce0a6598ea32537fabfd9cd2d0f389cfd023f319df5e05d8977b2fb6: exit status 1
vgo: stat github.com/yvasiyarov/gorelic@a9bba5b9ab508a086f9a12b8c51fab68478e2128: git fetch --depth=1 https://github.com/yvasiyarov/gorelic a9bba5b9ab508a086f9a12b8c51fab68478e2128 in /home/pwaller/.local/src/v/cache/vcswork/3a3ef5c2da0642f123b8cb28aa914d49f2fa618e45adddfd40b821560787999f: exit status 1
vgo: stat github.com/yvasiyarov/newrelic_platform_go@b21fdbd4370f3717f3bbd2bf41c223bc273068e6: git fetch --depth=1 https://github.com/yvasiyarov/newrelic_platform_go b21fdbd4370f3717f3bbd2bf41c223bc273068e6 in /home/pwaller/.local/src/v/cache/vcswork/b244b25b60c150ec4f0f7d9f8654f09a9abc8fa82d57b52ba48ed048cb02fd3d: exit status 1
vgo: stat google.golang.org/appengine@12d5545dc1cfa6047a286d5e853841b6471f4c19: git fetch --depth=1 https://github.com/golang/appengine 12d5545dc1cfa6047a286d5e853841b6471f4c19 in /home/pwaller/.local/src/v/cache/vcswork/606c3f865bae05dafcee0ab4f2fcffe95f20e2ab6a3430cd814a17d4baa50fe2: exit status 1
vgo: stat google.golang.org/grpc@d3ddb4469d5a1b949fc7a7da7c1d6a0d1b6de994: git fetch --depth=1 https://github.com/grpc/grpc-go d3ddb4469d5a1b949fc7a7da7c1d6a0d1b6de994 in /home/pwaller/.local/src/v/cache/vcswork/53ab5f2f034ba42de32f909aa45670cf730847987f38664c4052b329152ad727: exit status 1
vgo: stat rsc.io/letsencrypt@e770c10b0f1a64775ae91d240407ce00d1a5bdeb: git fetch --depth=1 https://github.com/rsc/letsencrypt e770c10b0f1a64775ae91d240407ce00d1a5bdeb in /home/pwaller/.local/src/v/cache/vcswork/092b467b1b246260ef830bdfbfaefa977c63414385ce3b90695546a830e4e77b: exit status 1
vgo: finding gopkg.in/square/go-jose.v1 v1.0.0-20160329203311-40d457b43924
vgo: finding google.golang.org/api v0.0.0-20160322025152-9bf6e6e569ff
vgo: finding github.com/aws/aws-sdk-go v0.0.0-20160708000820-90dec2183a5f
vgo: finding github.com/coreos/go-systemd v0.0.0-20151104194251-b4a58d95188d
vgo: finding github.com/aws/aws-sdk-go v1.4.22
vgo: finding github.com/Sirupsen/logrus v0.11.0
vgo: finding github.com/Microsoft/hcsshim v0.5.9
vgo: finding github.com/Microsoft/go-winio v0.3.8
vgo: downloading github.com/docker/docker v1.13.1
vgo: downloading github.com/docker/distribution v0.0.0-20161216195125-28602af35ace
vgo: downloading github.com/docker/go-connections v0.0.0-20170204000113-ecb4cb2dd420
vgo: downloading github.com/Sirupsen/logrus v0.11.0
vgo: downloading golang.org/x/net v0.0.0-20160726221601-6a513affb38d
vgo: downloading github.com/pkg/errors v0.0.0-20160613021747-01fa4104b9c2
vgo: resolving import "github.com/docker/go-units"
vgo: finding github.com/docker/go-units (latest)
vgo: adding github.com/docker/go-units v0.3.3
vgo: resolving import "github.com/opencontainers/runc/libcontainer/user"
vgo: stat github.com/mrunalp/fileutils@ed869b029674c0e9ce4c0dfa781405c2d9946d08: git fetch --depth=1 origin ed869b029674c0e9ce4c0dfa781405c2d9946d08 in /home/pwaller/.local/src/v/cache/vcswork/302c9e0d4c9eb9d02d5aab1990f5600ac4d185a0b20bbe863d2288078a915775: exit status 1
vgo: stat github.com/seccomp/libseccomp-golang@84e90a91acea0f4e51e62bc1a75de18b1fc0790f: git fetch --depth=1 https://github.com/seccomp/libseccomp-golang 84e90a91acea0f4e51e62bc1a75de18b1fc0790f in /home/pwaller/.local/src/v/cache/vcswork/ea142ed8c34e9a56ddedd65452487d29c5612d615c17f06d66852712fc41d44c: exit status 1
vgo: stat github.com/syndtr/gocapability@db04d3cc01c8b54962a58ec7e491717d06cfcc16: git fetch --depth=1 https://github.com/syndtr/gocapability db04d3cc01c8b54962a58ec7e491717d06cfcc16 in /home/pwaller/.local/src/v/cache/vcswork/9325046eaf69fe74b7d765df9e31c31ceb3f57013ea20b2db645b2343184287d: exit status 1
vgo: stat github.com/vishvananda/netlink@1e2e08e8a2dcdacaae3f14ac44c5cfa31361f270: git fetch --depth=1 https://github.com/vishvananda/netlink 1e2e08e8a2dcdacaae3f14ac44c5cfa31361f270 in /home/pwaller/.local/src/v/cache/vcswork/4ff59ae4f2cec838b4e8f5ba43b230f5428d35f58a9298babb0097f083d77331: exit status 1
vgo: stat github.com/golang/protobuf@18c9bb3261723cd5401db4d0c9fbc5c3b6c70fe8: git fetch --depth=1 https://github.com/golang/protobuf 18c9bb3261723cd5401db4d0c9fbc5c3b6c70fe8 in /home/pwaller/.local/src/v/cache/vcswork/a20e27c072e7660590311c08d0311b908b66675cb634b0caf7cbb860e5c3c705: exit status 1
vgo: stat github.com/urfave/cli@d53eb991652b1d438abdd34ce4bfa3ef1539108e: git fetch --depth=1 origin d53eb991652b1d438abdd34ce4bfa3ef1539108e in /home/pwaller/.local/src/v/cache/vcswork/0ed98460fb5750d57cdf7c0d08d87fff9b4edc179519b099ced12249c66ab69c: exit status 1
vgo: stat github.com/containerd/console@2748ece16665b45a47f884001d5831ec79703880: git fetch --depth=1 origin 2748ece16665b45a47f884001d5831ec79703880 in /home/pwaller/.local/src/v/cache/vcswork/a97da8eae0e8068197968e0ef6767b85aa97d8627bfcf35df2d7431b8b8f7bec: exit status 1
vgo: finding github.com/opencontainers/runc (latest)
vgo: adding github.com/opencontainers/runc v1.0.0-rc5
vgo: finding github.com/opencontainers/runc v1.0.0-rc5
vgo: stat github.com/mrunalp/fileutils@ed869b029674c0e9ce4c0dfa781405c2d9946d08: git fetch --depth=1 https://github.com/mrunalp/fileutils ed869b029674c0e9ce4c0dfa781405c2d9946d08 in /home/pwaller/.local/src/v/cache/vcswork/302c9e0d4c9eb9d02d5aab1990f5600ac4d185a0b20bbe863d2288078a915775: exit status 1
vgo: stat github.com/seccomp/libseccomp-golang@84e90a91acea0f4e51e62bc1a75de18b1fc0790f: git fetch --depth=1 https://github.com/seccomp/libseccomp-golang 84e90a91acea0f4e51e62bc1a75de18b1fc0790f in /home/pwaller/.local/src/v/cache/vcswork/ea142ed8c34e9a56ddedd65452487d29c5612d615c17f06d66852712fc41d44c: exit status 1
vgo: stat github.com/syndtr/gocapability@db04d3cc01c8b54962a58ec7e491717d06cfcc16: git fetch --depth=1 https://github.com/syndtr/gocapability db04d3cc01c8b54962a58ec7e491717d06cfcc16 in /home/pwaller/.local/src/v/cache/vcswork/9325046eaf69fe74b7d765df9e31c31ceb3f57013ea20b2db645b2343184287d: exit status 1
vgo: stat github.com/vishvananda/netlink@1e2e08e8a2dcdacaae3f14ac44c5cfa31361f270: git fetch --depth=1 https://github.com/vishvananda/netlink 1e2e08e8a2dcdacaae3f14ac44c5cfa31361f270 in /home/pwaller/.local/src/v/cache/vcswork/4ff59ae4f2cec838b4e8f5ba43b230f5428d35f58a9298babb0097f083d77331: exit status 1
vgo: stat github.com/golang/protobuf@18c9bb3261723cd5401db4d0c9fbc5c3b6c70fe8: git fetch --depth=1 https://github.com/golang/protobuf 18c9bb3261723cd5401db4d0c9fbc5c3b6c70fe8 in /home/pwaller/.local/src/v/cache/vcswork/a20e27c072e7660590311c08d0311b908b66675cb634b0caf7cbb860e5c3c705: exit status 1
vgo: stat github.com/urfave/cli@d53eb991652b1d438abdd34ce4bfa3ef1539108e: git fetch --depth=1 https://github.com/urfave/cli d53eb991652b1d438abdd34ce4bfa3ef1539108e in /home/pwaller/.local/src/v/cache/vcswork/0ed98460fb5750d57cdf7c0d08d87fff9b4edc179519b099ced12249c66ab69c: exit status 1
vgo: stat github.com/containerd/console@2748ece16665b45a47f884001d5831ec79703880: git fetch --depth=1 https://github.com/containerd/console 2748ece16665b45a47f884001d5831ec79703880 in /home/pwaller/.local/src/v/cache/vcswork/a97da8eae0e8068197968e0ef6767b85aa97d8627bfcf35df2d7431b8b8f7bec: exit status 1
vgo: finding golang.org/x/sys v0.0.0-20170901181214-7ddbeae9ae08
vgo: finding github.com/sirupsen/logrus v0.0.0-20170713114250-a3f95b5c4235
vgo: finding github.com/pkg/errors v0.8.0
vgo: finding github.com/opencontainers/selinux v1.0.0-rc1
vgo: finding github.com/opencontainers/runtime-spec v1.0.0
vgo: finding github.com/godbus/dbus v0.0.0-20151105175453-c7fdd8b5cd55
vgo: finding github.com/docker/go-units v0.2.0
vgo: finding github.com/cyphar/filepath-securejoin v0.2.1
vgo: finding github.com/coreos/go-systemd v0.0.0-20161114122254-48702e0da86b
vgo: finding github.com/docker/go-units v0.3.3
vgo: downloading github.com/docker/go-units v0.3.3
vgo: downloading github.com/opencontainers/runc v1.0.0-rc5
vgo: downloading golang.org/x/sys v0.0.0-20170901181214-7ddbeae9ae08
vgo: downloading github.com/pkg/errors v0.8.0

real	8m46.055s
user	0m40.632s
sys	0m14.428s
```
