日志
=====

导入日志库

    import "lg"

基本上等同于系统`log`库，在上面封装了level并且增加了`Lrelativefile`，只打印相对路径

    lg.SetFlags(lg.LstdFlags | lg.Lrelativefile)


日志可以分等级

    lg.Debug("debug info")
    lg.Infof("debug info")
    lg.Warnf("warn info")
    lg.Errorf("error")
    lg.Panicf("panic")
    lg.Fatalf("fatail") # 会退出程序

只查看warn及其以上登录的日志

    lg.SetLevel(lg.Lwarn)

也可以这样

    lg.SetLevelByName("warn")

设置日志输入到标准输出

    lg.SetOutput(os.Stdout)

设置日志前缀

    lg.SetPrefix("[zwc] ")


日志文件rotate

    w := lg.NewFileWriter()
    w.Open("/data/logs/api_user.log")
    w.MaxLines = 100000 # 日志文件写足10万行后就rotate，默认为0
    w.MaxDays = 7 # 最大保存日志天数，默认为7
    lg.SetOutput(w) # 设置日志到文件
