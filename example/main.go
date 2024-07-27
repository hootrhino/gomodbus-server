// Copyright (C) 2024 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"github.com/hootrhino/gomodbus-server"
	logrus "github.com/sirupsen/logrus"
	"log"
	"time"
)

func main() {
	server := mbserver.NewServerWithContext(context.Background())
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	server.SetLogger(logger)

	err := server.ListenTCP("127.0.0.1:1502")
	if err != nil {
		log.Printf("%v\n", err)
	}
	defer server.Close()

	// Wait forever
	for {
		time.Sleep(1 * time.Second)
	}
}
