package syslog

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/smithy-go/ptr"
	"github.com/c2pc/go-pkg/v2/utils/app_data"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/gin-gonic/gin"
)

var data Data

func init() {
	data = DataFromSystem()
}

type Data struct {
	ClientIP    string     `json:"client_ip"`
	ClientHost  string     `json:"client_host"`
	ClientPort  string     `json:"client_port"`
	ServerIP    string     `json:"server_ip"`
	ServerHost  string     `json:"server_host"`
	ServerPort  string     `json:"server_port"`
	ServerProto string     `json:"server_proto"`
	ServerMac   string     `json:"server_mac"`
	UserLogin   string     `json:"user_login"`
	StartTime   *time.Time `json:"start_time,omitempty"`
}

func DataFromRequest(c *gin.Context) Data {
	r := data

	r.StartTime = ptr.Time(time.Now())
	r.ClientIP = c.ClientIP()
	r.ServerProto = c.GetHeader("X-Forwarded-Proto")

	remoteAddr := c.Request.RemoteAddr
	if host, port, err := net.SplitHostPort(remoteAddr); err == nil {
		r.ClientHost = host
		r.ClientPort = port
	}

	r.ServerHost = c.Request.Host
	r.ServerMac = macAddress()

	if addr := c.Request.Context().Value(http.LocalAddrContextKey); addr != nil {
		if tcpAddr, ok := addr.(*net.TCPAddr); ok {
			r.ServerIP = tcpAddr.IP.String()
			r.ServerPort = strconv.Itoa(tcpAddr.Port)
		}
	}

	return r
}

func DataFromSystem() Data {
	r := Data{}

	r.ServerMac = macAddress()
	interfaces, _ := app_data.Interfaces()
	func() {
		for _, i := range interfaces {
			r.ServerMac = macAddress()
			if i.Active {
				for _, ip := range i.Addresses {
					if ip.Family == "inet" {
						r.ServerIP = ip.IP.String()
						return
					}
				}
			}
		}
	}()

	return r
}

type contextKey string

var CtxDataKey = contextKey("syslog-data")

func CtxWithData(ctx context.Context, data Data) context.Context {
	return context.WithValue(ctx, CtxDataKey, data)
}

func CtxGetData(ctx context.Context) Data {
	v := ctx.Value(CtxDataKey)
	if v == nil {
		return data
	}

	d, ok := v.(Data)
	if !ok {
		return data
	}

	d.UserLogin, _ = mcontext.GetOpUserLogin(ctx)
	return d
}

func macAddress() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || len(iface.HardwareAddr) == 0 {
			continue
		}

		return iface.HardwareAddr.String()
	}
	return ""
}
