package main

import (
	"context"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/option"
)

func CreateSpannerClient(ctx context.Context, db string, o ...option.ClientOption) (*spanner.Client, error) {
	dataClient, err := spanner.NewClient(ctx, db, o...)
	if err != nil {
		return nil, err
	}

	return dataClient, nil
}

//func (s *SpannerService) ExactStalenessQuery(ctx context.Context, sql string) {
//	fmt.Printf("Start Query : %s\n", sql)
//	iter := s.sc.Single().WithTimestampBound(spanner.ExactStaleness(time.Second*15)).QueryWithStats(ctx, spanner.Statement{
//		SQL: sql,
//	})
//	defer iter.Stop()
//	for {
//		row, err := iter.Next()
//		if err == iterator.Done {
//			break
//		}
//		if err != nil {
//			panic(err)
//		}
//		var count int64
//		if err := row.ColumnByName("Count", &count); err != nil {
//			panic(err)
//		}
//		fmt.Printf("Count:%d\n", count)
//	}
//}
//
//func (s *SpannerService) PartitionedDML(ctx context.Context, sql string) (int64, error) {
//	defer func(n time.Time) {
//		d := time.Since(n)
//		fmt.Printf("PartitionedDML:Time: %v \n", d)
//	}(time.Now())
//
//	stmt := spanner.Statement{SQL: sql}
//	rowCount, err := s.sc.PartitionedUpdate(ctx, stmt)
//	if err != nil {
//		return 0, err
//	}
//
//	return rowCount, nil
//}
