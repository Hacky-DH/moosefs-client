package mfscli

const MFS_ROOT_ID = 1
const MFS_NAME_MAX = 255
const MFS_SYMLINK_MAX = 4096
const MFS_PATH_MAX = 1024

const FUSE_REGISTER_BLOB_ACL = "DjI1GAQDULI5d2YjA26ypc3ovkhjvhciTQVx3CS4nYgtBoUcsljiVpsErJENHaw0"

const ANTOAN_NOP = 0

// [msgid:32]
// send/recv:  ANTOAN_NOP(32,0) size(32,4) id(32,0)

const REGISTER_GETRANDOM uint8 = 1

// rcode==1: generate random blob
// CLTOMA:
//  rcode:8
// MATOCL:
//  randomblob:32B

const REGISTER_NEWSESSION uint8 = 2

// CLTOMA:
// rcode:8 version:32 ileng:32 info:ilengB pleng:32 path:plengB
// [ sessionid:32 [ metaid:64 ]] [ passcode:16B ]
// MATOCL:
// version:32 sessionid:32 metaid:64 sesflags:8 rootuid:32 rootgid:32
// mapalluid:32 mapallgid:32 mingoal:8 maxgoal:8 mintrashtime:32 maxtrashtime:32
// ( version >= 3.0.72 )

const REGISTER_RECONNECT uint8 = 3

// CLTOMA:
//  rcode:8 sessionid:32 version:32 [ metaid:64 ]
// MATOCL:
//  status:8

const REGISTER_TOOLS uint8 = 4

// CLTOMA:
//  rcode:8 sessionid:32 version:32
// MATOCL:
//  status:8

const REGISTER_NEWMETASESSION uint8 = 5

// CLTOMA:
//  rcode:8 version:32 ileng:32 info:ilengB [ sessionid:32 [ metaid:64 ]] [ passcode:16B ]
// MATOCL:
//  version:32 sessionid:32 metaid:64 sesflags:8 mingoal:8 maxgoal:8 mintrashtime:32
//  maxtrashtime:32 ( version >= 3.0.11 )
//  status:8

const REGISTER_CLOSESESSION uint8 = 6

// CLTOMA:
//  rcode:8 sessionid:32 [ metaid:64 ]
// MATOCL:
//  status:8

const ANTOAN_GET_VERSION = 10

// [msgid:32] version:32 strversion:string ( N*[ char:8 ] )

const CLTOMA_FUSE_REGISTER = 400

const CLTOMA_FUSE_STATFS = 402

// CLTOMA
// msgid:32 -
// MATOCL
// msgid:32 totalspace:64 availspace:64 freespace:64 trashspace:64
// sustainedspace:64 inodes:32

const CLTOMA_FUSE_ACCESS = 404

// CLTOMA
// msgid:32 inode:32 uid:32 gcnt:32 gcnt * [ gid:32 ] perm:16
// MATOCL
// msgid:32 status:8

const CLTOMA_FUSE_LOOKUP = 406

// CLTOMA
// msgid:32 inode:32 nleng:8 name:NAME uid:32 gcnt:32 gcnt * [ gid:32 ]
// MATOCL
// msgid:32 status:8
// msgid:32 inode:32 attr:ATTR lflags:16 [ protocolid:8 chunkid:64 version:32
// N * [ ip:32 port:16 cs_ver:32 labelmask:32 ] ] - (master and client both
// versions >= 3.0.40 - protocolid==2 ; chunk 0 data only for one-chunk files
// with unlocked chunk)

const CLTOMA_FUSE_MKNOD = 416

// CLTOMA
// msgid:32 inode:32 name:NAME type:8 mode:16 umask:16 uid:32 gcnt:32 gcnt * [
// gid:32 ] rdev:32
// MATOCL
// msgid:32 status:8
// msgid:32 inode:32 attr:ATTR

const CLTOMA_FUSE_MKDIR = 418

// CLTOMA
// msgid:32 inode:32 name:NAME mode:16 umask:16 uid:32 gcnt:32 gcnt * [ gid:32 ]
// copysgid:8
// MATOCL
// msgid:32 status:8
// msgid:32 inode:32 attr:ATTR

const CLTOMA_FUSE_UNLINK = 420

// CLTOMA
// msgid:32 inode:32 name:NAME uid:32 gcnt:32 gcnt * [ gid:32 ]
// MATOCL
// msgid:32 status:8

const CLTOMA_FUSE_RMDIR = 422

// CLTOMA
// msgid:32 inode:32 name:NAME uid:32 gcnt:32 gcnt * [ gid:32 ]
// MATOCL
// msgid:32 status:8

const CLTOMA_FUSE_RENAME = 424

// CLTOMA
// msgid:32 inode_src:32 name_src:NAME inode_dst:32 name_dst:NAME uid:32 gcnt:32
// gcnt * [ gid:32 ]
// MATOCL
// msgid:32 status:8
// msgid:32 inode:32 attr:ATTR

const CLTOMA_FUSE_LINK = 426

// CLTOMA
// msgid:32 inode:32 inode_dst:32 name_dst:NAME uid:32 gcnt:32 gcnt * [ gid:32 ]
// MATOCL
// msgid:32 status:8
// msgid:32 inode:32 attr:ATTR

const CLTOMA_FUSE_READDIR = 428

// CLTOMA
// msgid:32 inode:32 uid:32 gcnt:32 gcnt * [ gid:32 ] flags:8 maxentries:32
// nedgeid:64
// MATOCL
// msgid:32 status:8
// msgid:32 [ nedgeid:64 ] N*[ name:NAME inode:32 type:8 ]	- when
// GETDIR_FLAG_WITHATTR in flags is not set
// msgid:32 [ nedgeid:64 ] N*[ name:NAME inode:32 attr:ATTR ]	- when
// GETDIR_FLAG_WITHATTR in flags is set

const CLTOMA_FUSE_OPEN = 430

// CLTOMA
// msgid:32 inode:32 uid:32 gcnt:32 gcnt * [ gid:32 ] flags:8
// MATOCL
// msgid:32 status:8
// msgid:32 attr:ATTR

const CLTOMA_FUSE_READ_CHUNK = 432

// CLTOMA
// msgid:32 inode:32 chunkindx:32 chunkopflags:8
// MATOCL
// maxsize=4096
// msgid:32 status:8
// msgid:32 protocolid:8 length:64 chunkid:64 version:32 N * [ ip:32 port:16
// cs_ver:32 labelmask:32 ] (master and client both versions >= 3.0.10 -
// protocolid==2)

const CLTOMA_FUSE_WRITE_CHUNK = 434

// CLTOMA
// msgid:32 inode:32 chunkindx:32 chunkopflags:8
// MATOCL
// maxsize=4096
// msgid:32 status:8
// msgid:32 protocolid:8 length:64 chunkid:64 version:32 N * [ ip:32 port:16
// cs_ver:32 labelmask:32 ] (master and client both versions >= 3.0.10 -
// protocolid==2)

const CLTOMA_FUSE_WRITE_CHUNK_END = 436

// CLTOMA
// msgid:32 chunkid:64 inode:32 chunkindx:32 length:64 chunkopflags:8
// MATOCL
// msgid:32 status:8

const CLTOMA_FUSE_QUOTACONTROL = 476

// CLTOMA
// msgid:32 inode:32 qflags:8 - delete quota
// msgid:32 inode:32 qflags:8 sinodes:32 slength:64 ssize:64 srealsize:64
// hinodes:32 hlength:64 hsize:64 hrealsize:64 - set quota
// MATOCL
// msgid:32 qflags:8 graceperiod:32 sinodes:32 slength:64 ssize:64 srealsize:64
// hinodes:32 hlength:64 hsize:64 hrealsize:64 curinodes:32 curlength:64
// cursize:64 currealsize:64 (size = 93, version >= 3.0.9)
// msgid:32 status:8

const CLTOMA_INFO = 510

// MATOCL:
// version:32 memusage:64 syscpu:64 usercpu:64 totalspace:64 availspace:64
// trashspace:64 trashnodes:32 sustainedspace:64 sustainednodes:32 allnodes:32
// dirnodes:32 filenodes:32 chunks:32 chunkcopies:32 tdcopies:32 laststore_ts:32
// laststore_duration:32 laststore_status:8 state:8 nstate:8 stable:8 sync:8
// leaderip:32 state_chg_time:32 meta_version:64 (size = 121,version >= 2.0.0)

const CLTOMA_QUOTA_INFO = 518

// MATOCL:
// quota_time_limit:32 N*[ inode:32 pleng:32 path:plengB graceperiod:32
// exceeded:8 qflags:8 stimestamp:32 sinodes:32 slength:64 ssize:64 srealsize:64
// hinodes:32 hlength:64 hsize:64 hrealsize:64 currinodes:32 currlength:64
// currsize:64 currrealsize:64 ]

const CLTOMA_SESSION_LIST = 508

// MATOCL:
// stats:16 N*[ sessionid:32 ip:32 version:32 openfiles:32 nsocks:8 expire:32
// ileng:32 info:ilengB pleng:32 path:plengB sesflags:8 rootuid:32
// rootgid:32 mapalluid:32 mapallgid:32 mingoal:8 maxgoal:8 mintrashtime:32
// maxtrashtime:32 stats * [ current_statdata:32 ] stats * [ last_statdata:32 ] ]
// - vmode = 2 (minsize = 188, valid since version 3.0.72)

const CLTOMA_SESSION_COMMAND = 526

// CLTOMA:
// commandid:8 sessionid:32
// commandid = 0 remove session

var ERROR_TABLE = []string{
	"OK",
	"Operation not permitted",
	"Not a directory",
	"No such file or directory",
	"Permission denied",
	"File exists",
	"Invalid argument",
	"Directory not empty",
	"Chunk lost",
	"Out of memory",
	"Index too big",
	"Chunk locked",
	"No chunk servers",
	"No such chunk",
	"Chunk is busy",
	"Incorrect register BLOB",
	"Operation not completed",
	"File not opened",
	"Write not started",
	"Wrong chunk version",
	"Chunk already exists",
	"No space left",
	"IO error",
	"Incorrect block number",
	"Incorrect size",
	"Incorrect offset",
	"Can't connect",
	"Incorrect chunk id",
	"Disconnected",
	"CRC error",
	"Operation delayed",
	"Can't create path",
	"Data mismatch",
	"Read-only file system",
	"Quota exceeded",
	"Bad session id",
	"Password is needed",
	"Incorrect password",
	"Attribute not found",
	"Operation not supported",
	"Result too large",
	"Entity not found",
	"Entity is active",
	"Chunkserver not present",
	"Waiting on lock",
	"Resource temporarily unavailable",
	"Interrupted system call",
	"Operation canceled",
	"No such file or directory (not cacheable)",
	"Operation not permitted (mfs admin only)",
	"Class name already in use",
	"Maximum number of classes reached",
	"No such class",
	"Class in use",
	"Unknown MFS error",
}

func MFSStrerror(code uint8) string {
	max := uint8(len(ERROR_TABLE) - 1)
	if code > max {
		code = max
	}
	return ERROR_TABLE[code]
}
