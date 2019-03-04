package mfscli

/*
MIT License

Copyright (c) 2019 DHacky
*/

import "time"

const TCP_RETRY_TIMES = 3
const TCP_CONNECT_TIMEOUT = 30 * time.Second
const TCP_RW_TIMEOUT = time.Minute
const MASTER_HEARTBEAT_INTERVAL = 5 * time.Second

const MFS_ROOT_ID = 1
const MFS_NAME_MAX = 255
const MFS_SYMLINK_MAX = 4096
const MFS_PATH_MAX = 1024
const MIN_SPECIAL_INODE = 0x7FFFFFF0

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

// CHUNKSERVER <-> CLIENT/CHUNKSERVER

const CLTOCS_READ = 200

// chunkid:64 version:32 offset:32 size:32
// protocolid:8 chunkid:64 version:32 offset:32 size:32 (both versions >= 1.7.32)

const CSTOCL_READ_STATUS = 201

// chunkid:64 status:8

const CSTOCL_READ_DATA = 202

// chunkid:64 blocknum:16 offset:16 size:32 crc:32 size*[ databyte:8 ]

const CLTOCS_WRITE = 210

// chunkid:64 version:32 N*[ ip:32 port:16 ]
// protocolid:8 chunkid:64 version:32 N*[ ip:32 port:16 ] (both versions >= 1.7.32)

const CSTOCL_WRITE_STATUS = 211

// chunkid:64 writeid:32 status:8

const CLTOCS_WRITE_DATA = 212

// chunkid:64 writeid:32 blocknum:16 offset:16 size:32 crc:32 size*[ databyte:8 ]

const CLTOCS_WRITE_FINISH = 213

// chunkid:64 version:32

//ANY <-> CHUNKSERVER

const ANTOCS_GET_CHUNK_CHECKSUM = 300

// chunkid:64 version:32

const CSTOAN_CHUNK_CHECKSUM = 301

// chunkid:64 version:32 checksum:32
// chunkid:64 version:32 status:8

const ANTOCS_GET_CHUNK_CHECKSUM_TAB = 302

// chunkid:64 version:32

const CSTOAN_CHUNK_CHECKSUM_TAB = 303

// maxsize=4108
// chunkid:64 version:32 1024*[checksum:32]
// chunkid:64 version:32 status:8

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

const CLTOMA_FUSE_GETATTR = 408

// CLTOMA
// msgid:32 inode:32
// msgid:32 inode:32 uid:32 gid:32 - version <= 1.6.27
// msgid:32 inode:32 opened:8 uid:32 gid:32
// MATOCL
// msgid:32 status:8
// msgid:32 attr:ATTR

const CLTOMA_FUSE_SETATTR = 410

// CLTOMA
// msgid:32 inode:32 opened:8 uid:32 gcnt:32 gcnt * [ gid:32 ] setmask:8
// attrmode:16 attruid:32 attrgid:32 attratime:32 attrmtime:32 sugidclearmode:8
// - version < 3.0.93
// msgid:32 inode:32 opened:8 uid:32 gcnt:32 gcnt * [ gid:32 ] setmask:8
// attrmode:16 attruid:32 attrgid:32 attratime:32 attrmtime:32 winattr:8
// sugidclearmode:8
// MATOCL
// msgid:32 status:8
// msgid:32 attr:ATTR

const CLTOMA_FUSE_READLINK = 412

// CLTOMA
// msgid:32 inode:32
// MATOCL
// msgid:32 status:8
// msgid:32 length:32 path:lengthB

const CLTOMA_FUSE_SYMLINK = 414

// CLTOMA
// msgid:32 inode:32 name:NAME length:32 path:lengthB uid:32 gid:32 - version <
// 2.0.0
// msgid:32 inode:32 name:NAME length:32 path:lengthB uid:32 gcnt:32 gcnt * [
// gid:32 ]
// MATOCL
// msgid:32 status:8
// msgid:32 inode:32 attr:ATTR

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
//   cs_ver:32 labelmask:32 ] (master and client both versions >= 3.0.10 -
//   protocolid==2)

const CLTOMA_FUSE_WRITE_CHUNK = 434

// CLTOMA
// msgid:32 inode:32 chunkindx:32 chunkopflags:8
// MATOCL
// maxsize=4096
// msgid:32 status:8
// msgid:32 protocolid:8 length:64 chunkid:64 version:32 N * [ ip:32 port:16
//   cs_ver:32 labelmask:32 ] (master and client both versions >= 3.0.10 -
//   protocolid==2)

const CLTOMA_FUSE_WRITE_CHUNK_END = 436

// CLTOMA
// msgid:32 chunkid:64 inode:32 chunkindx:32 length:64 chunkopflags:8
// MATOCL
// msgid:32 status:8

const CLTOMA_FUSE_APPEND_SLICE = 438

// msgid:32 flags:8 inode:32 srcinode:32 slice_from:32
// slice_to:32 uid:32 gcnt:32 gcnt * [ gid:32 ]
// msgid:32 status:8

const CLTOMA_FUSE_CHECK = 440

// CLTOMA
// msgid:32 inode:32 chunkindx:32 (version >= 3.0.26)
// MATOCL
// maxsize=1000
// msgid:32 status:8 (common)
// msgid:32 12*[ chunks:32 ] - 0 copies, 1 copy, 2 copies, ..., 10+ copies,
//   'empty' copies (version >= 3.0.30 and no chunkindx)
// msgid:32 chunkid:64 version:32 N*[ ip:32 port:16 type:8 ]
//   (version >= 3.0.26 and chunkindx present)

const CLTOMA_FUSE_GETTRASHTIME = 442

// CLTOMA
// msgid:32 inode:32 gmode:8
// MATOCL
// maxsize=100000
// msgid:32 status:8
// msgid:32 tdirs:32 tfiles:32 tdirs*[ trashtime:32 dirs:32 ]
//   tfiles*[ trashtime:32 files:32 ]

const CLTOMA_FUSE_SETTRASHTIME = 444

// CLTOMA
// msgid:32 inode:32 uid:32 trashtimeout:32 smode:8
// MATOCL
// msgid:32 status:8
// msgid:32 changed:32 notchanged:32 notpermitted:32

const CLTOMA_FUSE_GETSCLASS = 446

// CLTOMA
// msgid:32 inode:32 gmode:8
// MATOCL
// maxsize=100000
// msgid:32 status:8
// msgid:32 gdirs:8 gfiles:8 gdirs*[ goal:8 dirs:32 | 0xFF:8
//   storage_class:NAME ] gfiles*[ goal:8 files:32 | 0xFF storage_class:NAME ]
//   (version >= 3.0.75)

const CLTOMA_FUSE_SETSCLASS = 448

// CLTOMA
// msgid:32 inode:32 uid:32 goal:8 smode:8 (any version)
// msgid:32 inode:32 uid:32 labelscnt:8 smode:8 labelscnt *
//   [ MASKORGROUP * [ labelmask:32 ] ] (version >= 2.1.0)
// msgid:32 inode:32 uid:32 zero:8 smode:8 create_mode:8 arch_delay:16
//   create_labelscnt:8 keep_labelscnt:8 arch_labelscnt:8 create_labelscnt *
//   [ MASKORGROUP * [ labelmask:32 ] ] keep_labelscnt * [ MASKORGROUP * [
//   labelmask:32 ] ] arch_labelscnt * [ MASKORGROUP * [ labelmask:32 ] ]
//   (version >= 3.0.0)
// msgid:32 inode:32 uid:32 0xFF:8 smode:8 storage_class:NAME
//   (version >= 3.0.75 && (smode & SMODE_TMASK) != SMODE_EXCHANGE )
// msgid:32 inode:32 uid:32 0xFF:8 smode:8 old_storage_class:NAME
//   new_storage_class:NAME  (version >= 3.0.75 && (smode & SMODE_TMASK)
//   == SMODE_EXCHANGE )
// MATOCL
// msgid:32 status:8
// msgid:32 changed:32 notchanged:32 notpermitted:32 [quotaexceeded:32]

const CLTOMA_FUSE_GETTRASH = 450

// CLTOMA
// msgid:32 (version < 3.0.64)
// msgid:32 trash_id:32 (version >= 3.0.64)
// MATOCL
// msgid:32 status:8
// msgid:32 N*[ name:NAME inode:32 ]

const CLTOMA_FUSE_GETDETACHEDATTR = 452

// CLTOMA
// msgid:32 inode:32 [ dtype:8 ]
// MATOCL
// msgid:32 status:8
// msgid:32 attr:ATTR

const CLTOMA_FUSE_GETTRASHPATH = 454

// CLTOMA
// msgid:32 inode:32
// MATOCL
// msgid:32 status:8
// msgid:32 length:32 path:lengthB

const CLTOMA_FUSE_SETTRASHPATH = 456

// CLTOMA
// msgid:32 inode:32 length:32 path:lengthB
// MATOCL
// msgid:32 status:8

const CLTOMA_FUSE_UNDEL = 458

// msgid:32 inode:32
// msgid:32 status:8

const CLTOMA_FUSE_PURGE = 460

// msgid:32 inode:32
// msgid:32 status:8

const CLTOMA_FUSE_GETDIRSTATS = 462

// CLTOMA
// msgid:32 inode:32
// MATOCL
// msgid:32 status:8
// msgid:32 inodes:32 dirs:32 files:32 chunks:32 length:64 size:64 gsize:64

const CLTOMA_FUSE_TRUNCATE = 464

// CLTOMA
// msgid:32 inode:32 [ opened:8 ] uid:32 gid:32 length:64 (version < 2.0.0)
// msgid:32 inode:32 opened:8 uid:32 gcnt:32 gcnt * [ gid:32 ] length:64
//   (version >= 2.0.0/3.0.0)
// msgid:32 inode:32 flags:8 uid:32 gcnt:32 gcnt * [ gid:32 ] length:64
//   (version >= 2.0.89/3.0.25)
// MATOCL
// msgid:32 status:8
// msgid:32 attr:ATTR

const CLTOMA_FUSE_REPAIR = 466

// CLTOMA
// msgid:32 inode:32 uid:32 gid:32 - version < 2.0.0
// msgid:32 inode:32 uid:32 gcnt:32 gcnt * [ gid:32 ]
// MATOCL
// msgid:32 status:8
// msgid:32 notchanged:32 erased:32 repaired:32

const CLTOMA_FUSE_SNAPSHOT = 468

// CLTOMA
// msgid:32 inode:32 inode_dst:32 name_dst:NAME uid:32 gid:32 canoverwrite:8
//   (version <= 1.6.27)
// msgid:32 inode:32 inode_dst:32 name_dst:NAME uid:32 gid:32 smode:8 umask:16
//   (version > 1.6.27 and version < 2.0.0)
// msgid:32 inode:32 inode_dst:32 name_dst:NAME uid:32 gcnt:32 gcnt * [ gid:32 ]
//   smode:8 umask:16 (version >= 2.0.0)
// MATOCL
// msgid:32 status:8

const CLTOMA_FUSE_GETSUSTAINED = 470

// CLTOMA
// msgid:32
// MATOCL
// msgid:32 status:8
// msgid:32 N*[ name:NAME inode:32 ]

const CLTOMA_FUSE_GETEATTR = 472

// CLTOMA
// msgid:32 inode:32 gmode:8
// MATOCL
// maxsize=100000
// msgid:32 status:8
// msgid:32 eattrdirs:8 eattrfiles:8 eattrdirs*[ eattr:8 dirs:32 ]
//   eattrfiles*[ eattr:8 files:32 ]

const CLTOMA_FUSE_SETEATTR = 474

// CLTOMA
// msgid:32 inode:32 uid:32 eattr:8 smode:8
// MATOCL
// msgid:32 status:8
// msgid:32 changed:32 notchanged:32 notpermitted:32

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

const CLTOMA_FUSE_GETXATTR = 478

// CLTOMA
// msgid:32 inode:32 opened:8 uid:32 gid:32 nleng:8 name:nlengB mode:8
//   (version < 2.0.0)
// msgid:32 inode:32 nleng:8 name:nlengB mode:8 opened:8 uid:32 gcnt:32
//   gcnt * [ gid:32 ] (version >= 2.0.0)
//   empty name = list names
//   mode:
//    0 - get data
//    1 - get length only
// MATOCL
// maxsize=100000
// msgid:32 status:8
// msgid:32 vleng:32
// msgid:32 vleng:32 value:vlengB

const CLTOMA_FUSE_SETXATTR = 480

// CLTOMA
// msgid:32 inode:32 uid:32 gid:32 nleng:8 name:8[NLENG] vleng:32 value:8[VLENG]
//   mode:8 (version < 2.0.0)
// msgid:32 inode:32 nleng:8 name:8[NLENG] vleng:32 value:8[VLENG] mode:8
//   opened:8 uid:32 gcnt:32 gcnt * [ gid:32 ] (version >= 2.0.0)
//   mode:
//    0 - create or replace
//    1 - create only
//    2 - replace only
//    3 - remove
// MATOCL
// msgid:32 status:8

const CLTOMA_FUSE_CREATE = 482

// CLTOMA
// msgid:32 inode:32 name:NAME mode:16 uid:32 gid:32 (version < 2.0.0)
// msgid:32 inode:32 name:NAME mode:16 umask:16 uid:32 gcnt:32 gcnt * [ gid:32 ]
// (version >= 2.0.0)
// MATOCL
// msgid:32 status:8
// msgid:32 inode:32 attr:ATTR

const CLTOMA_FUSE_PARENTS = 484

// CLTOMA
// msgid:32 inode:32
// MATOCL
// msgid:32 status:8
// msgid:32 N*[ inode:32 ]

const CLTOMA_FUSE_PATHS = 486

// CLTOMA
// msgid:32 inode:32
// MATOCL
// msgid:32 status:8
// msgid:32 N*[ length:32 path:lengthB ]

const CLTOMA_FUSE_GETFACL = 488

// CLTOMA
// msgid:32 inode:32 acltype:8 opened:8 uid:32 gcnt:32 gcnt * [ gid:32 ]
//   (version < 3.0.92)
// msgid:32 inode:32 acltype:8 (version >= 3.0.92)
// MATOCL
// msgid:32 status:8
// msgid:32 userperm:16 groupperm:16 otherperm:16 mask:16 namedusers:16
//   namedgroups:16 namedusers * [ id:32 perm:16 ] namedgroups * [ id:32 perm:16 ]

const CLTOMA_FUSE_SETFACL = 490

// CLTOMA
// msgid:32 inode:32 uid:32 acltype:8 userperm:16 groupperm:16 otherperm:16
//   mask:16 namedusers:16 namedgroups:16 namedusers * [ id:32 perm:16 ]
//   namedgroups * [ id:32 perm:16 ]
// MATOCL
// msgid:32 status:8

const CLTOMA_FUSE_FLOCK = 492

// CLTOMA
// msgid:32 inode:32 reqid:32 owner:64 cmd:8
// MATOCL
// msgid:32 status:8

const CLTOMA_FUSE_POSIX_LOCK = 494

// CLTOMA
// msgid:32 inode:32 reqid:32 owner:64 pid:32 cmd:8 type:8 start:64 end:64
// MATOCL
// msgid:32 status:8 (cmd != POSIX_LOCK_CMD_GET || status != STATUS_OK)
// msgid:32 pid:32 type:8 start:64 end:64 (cmd == POSIX_LOCK_CMD_GET &&
//   status == STATUS_OK)

const CLTOMA_FUSE_ARCHCTL = 496

// CLTOMA
// msgid:32 inode:32 cmd:8 (cmd==ARCHCTL_GET)
// msgid:32 inode:32 cmd:8 uid:32 (cmd==ARCHCTL_SET or cmd==ARCHCTL_CLR)
// MATOCL
// msgid:32 status:8
// msgid:32 archchunks:64 notarchchunks:64 archinodes:32 partialinodes:32
//   notarchinodes:32 (cmd==ARCHCTL_GET)
// msgid:32 chunkschanged:64 chunksnotchanged:64 inodesnotpermitted:32
//   (cmd==ARCHCTL_SET or cmd==ARCHCTL_CLR)

const CLTOMA_FUSE_FSYNC = 498

// CLTOMA
// msgid:32 inode:32
// MATOCL
// msgid:32 status:8

const CLTOMA_SESSION_LIST = 508

// MATOCL:
// stats:16 N*[ sessionid:32 ip:32 version:32 openfiles:32 nsocks:8 expire:32
// ileng:32 info:ilengB pleng:32 path:plengB sesflags:8 rootuid:32
// rootgid:32 mapalluid:32 mapallgid:32 mingoal:8 maxgoal:8 mintrashtime:32
// maxtrashtime:32 stats * [ current_statdata:32 ] stats * [ last_statdata:32 ] ]
// - vmode = 2 (minsize = 188, valid since version 3.0.72)

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
