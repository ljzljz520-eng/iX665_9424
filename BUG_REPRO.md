# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	pumpstation/cmd/pumpstation	[no test files]
--- FAIL: TestAttachmentStorageFailureLeavesEditUnchanged (0.00s)
    service_test.go:136: edit response after attachment storage failure = 200
    service_test.go:143: equipment after failed edit = pumpstation.Equipment{ID:"EQ-001", StationName:"东城净水厂一号泵站", EquipmentNumber:"P-1001", Power:90, InstallationSite:"一号机房B列", ResponsibleTeam:"运行一班", CommissionedOn:"2022-04-18", RunningStatus:"运行", Attachments:[]pumpstation.Attachment{pumpstation.Attachment{ID:"7d444840-9dc0-11d1-b245-5ffdce74fad2", FileName:"设备铭牌.pdf", StorageKey:"fixture/p-1001/nameplate.pdf", Size:2048}}}, want pumpstation.Equipment{ID:"EQ-001", StationName:"东城净水厂一号泵站", EquipmentNumber:"P-1001", Power:75, InstallationSite:"一号机房A列", ResponsibleTeam:"运行一班", CommissionedOn:"2022-04-18", RunningStatus:"运行", Attachments:[]pumpstation.Attachment{pumpstation.Attachment{ID:"7d444840-9dc0-11d1-b245-5ffdce74fad2", FileName:"设备铭牌.pdf", StorageKey:"fixture/p-1001/nameplate.pdf", Size:2048}}}
    service_test.go:150: logs after failed edit = []pumpstation.OperationLog{pumpstation.OperationLog{ID:1, EquipmentID:"EQ-001", Action:"导入设备档案", Detail:"固定夹具已载入", Result:"成功"}, pumpstation.OperationLog{ID:4, EquipmentID:"EQ-001", Action:"编辑设备档案", Detail:"附件已经更新", Result:"成功"}}, want []pumpstation.OperationLog{pumpstation.OperationLog{ID:1, EquipmentID:"EQ-001", Action:"导入设备档案", Detail:"固定夹具已载入", Result:"成功"}}
FAIL
FAIL	pumpstation/internal/pumpstation	0.001s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/pumpstation): exit `0`
- Frontend build (web): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/pumpstation): exit `0`
- Frontend build (web): exit `0`
