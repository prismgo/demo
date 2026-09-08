#!/usr/bin/env bash
# Stream verbose test output into a bounded failure summary and a 12-line tail.
# The pipeline still needs `set -o pipefail`; this filter intentionally exits 0.
set -uo pipefail

awk '
    BEGIN {
        failre="--- FAIL|^FAIL|^# |panic:|FAIL|✗|✘|×|^Error|Error:|AssertionError|Expected|Received|[.]go:[0-9]+:[0-9]*:|[0-9]+ failed|not ok|Test Files.*failed|Tests .*failed|Unhandled|rejected [(]"
    }
    {
        total++
        tail[(total - 1) % 12] = $0
        if ($0 ~ failre) {
            failures++
            if (failures <= 60) hits[failures] = total ":" $0
        }
    }
    END {
        print "==== testfilter：原始 " total " 行 ===="
        if (failures) {
            print "-- 失败/报错行（原始行号）--"
            for (i=1; i<=failures && i<=60; i++) print hits[i]
            if (failures > 60) print "  …（更多失败行已省，需要细节看原始输出/落盘文件）"
        }
        print "-- 尾部汇总（末 12 行）--"
        start = total > 12 ? total - 11 : 1
        for (i=start; i<=total; i++) print tail[(i - 1) % 12]
        if (failures) print "==== 判读：FAIL（" failures " 处失败/报错信号）；最终判败以退出码为准 ===="
        else print "==== 判读：未见失败信号（疑 PASS）→ 仍以命令退出码 $? 确认 ===="
    }
'
