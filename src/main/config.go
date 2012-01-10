/*

Copyright 2011 Paul Ruane.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
    "bufio"
    "errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const globalConfigPath = "/etc/tmsu.conf"
const userConfigPath = "~/.config/tmsu.conf"

type Config struct {
    Databases []DatabaseConfig
}

type DatabaseConfig struct {
    Name string
    DatabasePath string
}

func GetSelectedDatabaseConfig() (*DatabaseConfig, error) {
    //TODO actually use selected rather than just default

    path, err := resolvePath("~/.tmsu/default.db")
    if err != nil { return nil, errors.New("Could not resolve default database path: " + err.Error()) }

    return &DatabaseConfig{ "default", path }, nil
}

func GetDatabaseConfig(name string) (*DatabaseConfig, error) {
    config, err := readConfig()
    if err != nil { return nil, err }

    for _, databaseConfig := range config.Databases {
        if databaseConfig.Name == name { return &databaseConfig, nil }
    }

    return nil, nil
}

func resolvePath(path string) (string, error) {
    if strings.HasPrefix(path, "~" + string(filepath.Separator)) {
        homeDirectory, err := os.Getenverror("HOME")
        if err != nil { return "", err }

        path = strings.Join([]string { homeDirectory, path[2:] }, string(filepath.Separator))
    }

    return path, nil
}

func readConfig() (*Config, error) {
    configPath, err := resolvePath(userConfigPath)
    if err != nil { return nil, err }

    file, err := os.Open(configPath)
    if err != nil { return nil, err }
    defer file.Close()

    reader := bufio.NewReader(file)

    databases := make([]DatabaseConfig, 0, 5)
    var database *DatabaseConfig

    for lineBytes, _, err := reader.ReadLine(); err == nil; lineBytes, _, err = reader.ReadLine() {
        line := string(lineBytes)
        trimmedLine := strings.TrimLeft(line, " \t")

        if len(trimmedLine) == 0 { continue }
        if strings.HasPrefix(trimmedLine, "#") { continue }

        var name, quotedValue string
        count, err := fmt.Sscanf(trimmedLine, "%s %s", &name, &quotedValue)
        if count < 2 { return nil, errors.New("Key and value must be specified.") }
        if err != nil { return nil, err }

        value, err := strconv.Unquote(quotedValue)
        if err != nil { return nil, errors.New("Configuration error: values must be quoted.") }

        switch name {
            case "database":
                if database != nil {
                    databases = append(databases, *database)
                }

                database = &DatabaseConfig{}
                database.Name = value
                if err != nil { return nil, err }
            case "path":
                path, err := resolvePath(value)
                if err != nil { return nil, err}

                database.DatabasePath = path
            default:
                return nil, errors.New("Unrecognised configuration element name '" + name + "'.");
        }
    }

    if database != nil {
        databases = append(databases, *database)
    }

    return &Config{ databases }, nil
}
