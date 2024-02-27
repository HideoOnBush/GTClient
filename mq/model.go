package mq

type Line struct {
	DocId          string `thrift:"docId,1" form:"docId" json:"docId" query:"docId"`
	DocVersion     int64  `thrift:"docVersion,2" form:"docVersion" json:"docVersion" query:"docVersion"`
	DocSeqNo       int64  `thrift:"docSeqNo,3" form:"docSeqNo" json:"docSeqNo" query:"docSeqNo"`
	DocPrimaryTerm int64  `thrift:"docPrimaryTerm,4" form:"docPrimaryTerm" json:"docPrimaryTerm" query:"docPrimaryTerm"`
	Source         string `thrift:"source,5" form:"source" json:"source" query:"source"`
	SourceIsCore   bool   `thrift:"sourceIsCore,6" form:"sourceIsCore" json:"sourceIsCore" query:"sourceIsCore"`
	SourceScene    string `thrift:"sourceScene,7" form:"sourceScene" json:"sourceScene" query:"sourceScene"`
	Target         string `thrift:"target,8" form:"target" json:"target" query:"target"`
	TargetIsCore   bool   `thrift:"targetIsCore,9" form:"targetIsCore" json:"targetIsCore" query:"targetIsCore"`
	TargetScene    string `thrift:"targetScene,10" form:"targetScene" json:"targetScene" query:"targetScene"`
	Dependence     string `thrift:"dependence,11" form:"dependence" json:"dependence" query:"dependence"`
}

type ConfMq struct {
	User string
	Pwd  string
	Addr string
	Port string
}
