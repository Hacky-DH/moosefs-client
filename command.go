package mfscli

const FUSE_REGISTER_BLOB_ACL = "DjI1GAQDULI5d2YjA26ypc3ovkhjvhciTQVx3CS4nYgtBoUcsljiVpsErJENHaw0"

const ANTOAN_NOP = 0

// [msgid:32]
// send/recv:  ANTOAN_NOP(32,0) size(32,4) id(32,0)

const REGISTER_NEWSESSION = 2

// CLTOMA:
// rcode:8 version:32 ileng:32 info:ilengB pleng:32 path:plengB
// [ sessionid:32 [ metaid:64 ]] [ passcode:16B ]
// MATOCL:
// version:32 sessionid:32 metaid:64 sesflags:8 umask:16 rootuid:32 rootgid:32
// mapalluid:32 mapallgid:32 mingoal:8 maxgoal:8 mintrashtime:32 maxtrashtime:32
// ( version >= 3.0.72 )

const REGISTER_TOOLS = 4

// CLTOMA:
//  rcode:8 sessionid:32 version:32
// MATOCL:
//  status:8

const ANTOAN_GET_VERSION = 10

// [msgid:32] version:32 strversion:string ( N*[ char:8 ] )

const CLTOMA_FUSE_REGISTER = 400

const CLTOMA_INFO = 510

const CLTOMA_QUOTA_INFO = 518

// quota_time_limit:32 N*[ inode:32 pleng:32 path:plengB [graceperiod:32]
// exceeded:8 qflags:8 stimestamp:32 sinodes:32 slength:64 ssize:64 sgoalsize:64
// hinodes:32 hlength:64 hsize:64 hgoalsize:64 currinodes:32 currlength:64
// currsize:64 currgoalsize:64 ]
