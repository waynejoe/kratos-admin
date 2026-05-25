package adminmodel

import pb "kratos-admin/pb/admin/v1"

func NewPageInfo(index, size int32, total int64, isEnd bool) *pb.PageInfo {
	return &pb.PageInfo{Index: index, Size: size, Total: int32(total), IsEnd: isEnd}
}
