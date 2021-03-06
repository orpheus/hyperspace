# Hyperspace

## Go far, move fast.

### Hyperledger Command Center

Hyperspace is a hyperledger command center that allows you to dynamically
spawn, create, destroy, and alter hyperledger nodes. It maintains a
memorybank of configurations and allows you to create, read, edit, and
delete, store, and version any configuration required to spawn a production ready network.

## Ideas
- create multiple network directories and have the configs and binaries
  isolated to each netdir
- dynamically create base configurations needed to start up a network

## Command Scripts Path Problem
- currently my paths are all over the place just to get things to
  work...
- `createCortium.sh` sets the working directory so I can run it
  anywhere, and `organizations/cpp-generate.sh` does the same thing so
  that it's path's can correctly reference the cpp-templates 
- I need to make it so I can run these commands anywhere without having
  to hack the `pwd` 
- in `cmd/cryptogen/main.go` I use an absolute path on my own directory
  structure to get it to correctly reference my config files and their
  relative paths... this needs to change as well

## Prerequisites

Hyperspace needs access to the fabric and fabric-ca code directly so it
can dynamically build the cmd binaries.

For running locally, pull down `fabric` and `fabric-ca` and place the
following scrips in the respective paths.


1. add `buildCmd.sh` and `buildNodes.sh` to `fabric/scripts/`

2. add `fabric/scripts` and `/fabric/bin` to `$PATH`

3. run `buildCmd.sh -t=ALL`

**!: ** currently for these build scripts to work, you need to call them
in the `fabric` root dir... toDo: add env vars

## Guide

### Make Crypto material for organizations
To create the crypto material, you need organization configurations. See
`/organizations/*` for examples. 

1. under `/cryptogen/`, configure the `cryptogen.yaml` to point to
   configuration paths for organization nodes (peers and orderers). By
   default they point to the config files found under `/organizations`.
   Generate the crypto material by running `go run
   cmd/cryptogen/main.go` or just build the binary and call it.

### Create the Genesis Block

To create the Genesis Block we create a consortium. This is what the
`configtxgen` binary does for us. You can see the config flags this
commander takes under `fabric/cmd/configtxgen` on the first line of the
`main` function.

Other than needing a `profile` `channelID` and `outputBlock`, the
configtxgen will look for a `configtx.yaml` configuration file. You can
specify this path via `configPath` flag or it'll look at the
`FABRIC_CFG_PATH` env variable. So you can either pass in

1, `configtxgen ..... -configPath "../path/to/cfg"`
2. `FABRIC_CFG_PATH="../path/to/cfg" configtxgen ,,,,`
3. add `export FABRIC_CFG_PATH="../path/to/cfg"` to your bash profile
4. run `export FABRIC_CFG_PATH="../path/to/cfg"` in current shell

**!:** the `configtx.yaml` file has to have a Profile that matches the
`-profile` you passed in to the `createConsortium.sh` script

**!:** then inside the `configtx.yaml` you need to point the
organization's msp directories to the correct cryptogen path which for
hyperspace is currently under `/cryptogen/organization/{my_orgs}`

**!:** make sure that the orgs and orderers found in the `configtx`
match the number of `orgs` in peers and orderers found under
`/cryptogen/organizations/{my_orgs}/**`

> CURRENTLY THE SYSTEM GENESIS BLOCK IS GENERATED RELATIVE TO WHERE YOU
CALL THE BASH SCRIPT SO CALL IT IN THE /CMDSCRIPTS/ DIR SO IT GETS
OUTPUT IN THE CORRECT PLACE

### SPAWNING AN ORDERER NODE
  - very important to note that the `orderer` does not let you set a
    config path with a flag. it looks for `orderer.yaml` in the default
    config paths which are `.` and `/env/hyperledger/fabric/` or via the
    env car `FABRIC_CFG_PATH`
  - also note: I copied the `configtx.yaml` to create the genesis block
    from `fabric-samples/test-network/configtx/configtx.yaml` and NOT
    from `fabric/sampleConfigs/configtx.yaml` OR `/fabric-samples/config/configtx.yaml`
    - `cp fabric-samples/test-network/configtx/configtx.yaml
      hyperspace/configtxgen/configtx.yaml`
  - NOTE: that `fabric/sampleConfigs` mirrors `fabric-samples/config`
    - verify that they're the same
  
  - so to spawn an `orderer` node I need to make sure the `orderer.yaml`
    is in the directory I called the binary from or add it to `FABRIC_CFG_PATH`
    - toDo: update orderer to accept custom config path
  - what I just did to get it to work was to copy `fabric-samples/config` right into `hyperspace`
    - then I moved just the `orderer.yaml` into the cmdscripts so the
      config could find it

  - it ran but then crashed because it was trying to access a directory
    that didn't exist
    - in the `orderer.yaml` I changed the `/var/**` dirs to `../orderer`
      (a local dir I temporarily created)
      - This WORKED

  - in `cmdscripts/` run `./spawn-orderer.sh`

### Spawning an orderer node from scratch via a static config
  - from now on, just make sure to call all cmdscripts from the
    cmdscript directory, think of it as your cmd-center

  - running `network.sh` will clean all the artifacts, generate the
    crypto, create a consortium, and spawn a node
    - eventually I'll have to check for existing material and only
      manually clean before spawn

### 12/14/2020 - Refactoring to Main Script Configuration Network Start

- cmdscripts now take a $1 -> $NETWORK argument so they can build the
  necessary artifacts in correct directories
- I moved the `cpp-generate.sh` script back to live under
  `{network}/organizations` because it was relative to that network. It
  was looking for files also belonging to the organizations folder so it
  made sense to move and keep it there. To get it to work correctly, I
  had to previously change the paths to point to the crypto material I
  generated locally and today I made it so I could pass down a BASE_PATH
  from `create-consoritum.sh` so that it would run succ. from the
  cmdcenter

New Directory Structure Is As So:

- Root
  - bin
  - cmdscripts
  - networks
    - [networkName]
      - cryptogen
      - organizations
      - configtxgen
      - nodes
      - hyperspace.yaml

This new directory structure makes it so I can generate other networks
and have all their material and configuration isolated

Things I forgot and needed to remind myself

  - `configtx.yaml` under configtxgen/ is needed by the `configtxgen`
    binary to create the consortiums and the genesis block
  - the orderer/ directory got moved to be the ledger/ directory and is
    now found under a specific orderer
    nodes/orderers/{orderer_01}/ledger
    - this is generated by the `orderer.yaml` file also found within a
      specific orderer's directory
      /nodes/orderers/{orderer_01}/orderer.yaml
    - the `orederer.yaml` is used by the `orderer` binary and IS NEEDED
      by this binary.
      - this binary doesn't accept config paths, so you need to hack
        fabric or hyperspace to make sure that the `orderer.yaml` gets
        accepted by the binary
      - this yaml is necessary for defining things like where to output
        the ledger data

## Hyperspace Configurations (yaml)

`hyperspace.yaml` is now used across the app

  - it makes it clear and distinct which configurations are part of this
    program
  - it makes configuration easier to work with in the go code
    - can now use one util function everywhere to check and read configs
  - the schema for the hyperspace configs are defined by their location
    in the network
    - their context determines their format
    - the go code will now what to do by where the `hyperspace.yaml`
      lives

## Next..
## 12/15/2020 - 
### Changes
- moved spawn-orderer.sh to deprecated folder
- moved scriptUtils to a utils folder
- now applying network-gate only to clean-artifacts
    - which is only automatically called by start-network with a hardcoded
     network name
- passing in -n (network) flag to scripts now (make-cryp & create-consort)
    - doing this now because we're no longer network gating which was very
     hardcoded anyways. plus, the network is in the main go script so it
      makes sense to pass that down and have that be more dynamic
      - had to add -n flag checks to while loops
- removed defaults for create-consort.sh (defined in script)

- so tired... 
- spent hours dealing with what I thought was an error...
    - turns out you need to set the stdout/err for golang's exec.Command
        - to see any output in the terminal
            - I thought the application was just hanging...
            
 `ps aux | grep orderer` to see running orderer process,  
 `kill -9 {PID}` grab PID from the orderer process
 
- I moved `orderer.yaml` to the cmdscript center to get it to work
    - now that I have the logging figured out, I'll test ways to get that config to work in it's own directory

- for the orderer hyperspace config, I added a placeholder: `${NETWORK_PATH}` to the env variables
    - which golang will parse and replace with the desired network path. Because I don't know which env vars
        - are paths at runtime, I have to parse+replace each one. 
            - this is compared to a config that I know is a path and can just append the network path
                - dynamic variables like this have to be dynamically parsed and replaced
                    **CREATE A KEY BANK OF PLACEHOLDER VARIABLES
                    - now I'm not using these because fabric is performing some kind of config path magic
 - this above comment is no longer relevant since paths are now relative in network hyperspace configs

- where you run everything... matters
- think about for local development saving all data to a ~/.hyperspace config

toDo:
1. ~~create a hyperspace config for configtxgen and create a main script to run and read from it to create the consortiums and the genesis block~~
    - DONE
2. ~~fix network main script~~
    ~~- orderer binary isn't getting the `orderer.yaml`~~
    - DONE
3. ~~spawn peers from main script~~
    ~~- update peer env vars~~
    ~~- make multiple binaries (manual)~~
     -DONE
4a. ~~convert configtxgen and cryptogen main scripts into packages that are used
 by the network main script~~ 
4b. ~~the network main script needs to create the ctg, cryp, and spawn the nodes
 (everything the start-network script is doing)~~ 
5. ~~add kill switch to main script which destroys processes and frees up ports~~

Idea: extra:
    - have network paths be more dynamics, hyperspace/networks/proj-network/proj-network-qa
        - create  different networks for different projects and different networks for different staging environments within that project
          
???? ~~How do I handle config paths??????~~ with command and control centers, the latter being done with an absolute base path to the project root
???? ~~How do I set the paths to the individual nodes without clashing?~~  -- DONE, send-commander
???? How do I make multiple binaries and put them with their folders?   


## 12-16-2020
1. cmdscripts can now be sent to cmdcenters anywhere in the network
   - for instance, the network main script is sending a spawn-node process
        - to the orderers directory and calling the binary IN the folder. Huge.
        
Think of cmdscripts as commanders or captain who command your fleet. They spread
out across the network issuing commands to different sectors of the fleet.

The MIND controlling all them is the Hyperspace binary. This binary is acts as the controller,
and sits or sets it's path to the control-center which is at the top of hyperspace where 
hyperspace is the network file systems. 

So now I need make sure paths are relative to where their commander called them.

2. in the peer hyperspace configs, I commented out the docker variables.
    - in the spawn-node.sh script I am word splitting on $START_CMD for the peers
        - "node start" command to run correctly
    - I made the TLS paths relative
3. I copied/pasted the `core.yaml` config from {network}/config (which I got from fabric-samples I believe)
    - in this I switched the ledger db path ( `fileSystemPath: ledger` ).
    - I also change the msp path (`mspConfigPath`) to the relative msp path
        - generated by the cryptogen binary
4. for the peer node config (`core.yaml`), the `ledger.snapshots.rootDir` has to be an absolute path
    - (I know, weird right?)
    - so I'm setting it to `/tmp/hyperspace/{peer-name}/ledger/snapshots`
5. I re-copied reset the /{network_name}/config directory with a copy from fabric-samples to 
    keep around original templates.
    - I then copied `core.yaml` over to each `peer-0x` directory and instead of changing the values
        in the `core.yaml`, I override them with the peer's `hyperspace.yaml`
    - in these hyperspace configs, note that each has two ports they listen on
        - I'm not yet sure what the operations ports is
        
## 12-16-2020
### PHASE 0.1.0-beta SPAWN TEST NETWORK WITH STATIC HYPERSPACE CONFIG COMPLETED
##### toDo: Cleanup
- ~~refactor cryptogen and configtxgen into modules~~ DONE 12/17/20
- ~~remove ./start-network.sh script and instead build and run a hyperspace binary~~ DONE 12/17/20
- ~~kill processes on SIGTERM~~ DONE 12/18/20

## 12-17-2020
- now assume you run commands from the control center
    - cmdscripts now needed to reference each other via a path
        relative to the control_center not the command_center
        
CONTROL_CENTER=/path/to/hyperspace_root_dir
COMMAND_CENTER=CONTROL_CENTER/cmdscripts

### ~~0.1.0 STATIC_CONFIG_ONE_SCRIPT_3NODES --> DONE -> DONE 12/18/20~~

## 0.2.0 CMDLINE INTERFACE
    - first dynamic implementations
    - generate network (Brahma)
    - start network by name (Shiva)
    - destroy network by name (Vishnu)
    - add node to network [by_name]
        - orderer
        - peer
    - add organization
    - create channel
    - install chaincode
    - CAs?...get the core functions scripted
 
DONE WHEN:
    - can create networks, orgs, and nodes (generate static resources)
    - can create channels and install chaincode
    - start and destroy networks by name
    
0.2.0_DONE -> PLAY WITH YOUR CREATION
    - build example networks with it
    - learn hyperledger inside and out

0.3.0...? Database or Docker? 
    - which first
    - Database will need a lot of work
        - filesystem management  
            - CRUD + versioning
        - apis
0.4.0...? Docker&Deploy or UI?
    - which first
    - UI will need a lot of work
        - I don't even want to think about it
    - deploying will need a lot of work
        - kubernetes integration
        - possibly that same filesystem management I need for Database or work

I would say, "save deployments till last and first build a usable application".
But making networks deployable before a UI or Database may be beneficial.
Or at least maybe after Database.
    - could I use hyperledger as my database to story it's own config? probably not safe
        - if net goes down, I lose my own filesystem memory
    - MEMORY_BANK

I could get to 0.2.0 and start building out the Hyperspace App I want.
Build the DB alongside it maybe... 

things to think about ~~~~
