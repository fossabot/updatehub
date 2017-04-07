/*
 * UpdateHub
 * Copyright (C) 2017
 * O.S. Systems Sofware LTDA: contato@ossystems.com.br
 *
 * SPDX-License-Identifier:     GPL-2.0
 */

package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/UpdateHub/updatehub/client"
	_ "github.com/UpdateHub/updatehub/installmodes/copy"
	"github.com/UpdateHub/updatehub/metadata"
	"github.com/UpdateHub/updatehub/utils"
)

func main() {
	osFs := afero.NewOsFs()

	fm, err := metadata.NewFirmwareMetadata(firmwareMetadataDirPath, osFs, &utils.CmdLine{})
	if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	uh := &UpdateHub{
		state:            NewIdleState(),
		api:              client.NewApiClient("localhost:8080"),
		updater:          client.NewUpdateClient(),
		timeStep:         time.Minute,
		store:            osFs,
		firmwareMetadata: *fm,
	}

	uh.Controller = uh

	uh.LoadSettings()
	uh.StartPolling()

	d := NewDaemon(uh)
	d.Run()
}
