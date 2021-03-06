package geo

import (
  "database/sql/driver"
  "fmt"

  "gitlab.com/spacewalker/geotracker/internal/pkg/util"
)

// Point represents coordinates [longitude, latitude] of the geographic position.
type Point [2]float64

// Longitude returns longitude of the pont.
func (p *Point) Longitude() float64 {
  return p[0]
}

// Latitude returns latitude of the pont.
func (p *Point) Latitude() float64 {
  return p[1]
}

// PostgresPoint is a postgresql representation of Point.
type PostgresPoint Point

// Value returns value in format that satisfies driver.Driver interface.
func (p PostgresPoint) Value() (driver.Value, error) {
  return fmt.Sprintf("(%v,%v)", p[0], p[1]), nil
}

// Scan parses raw value retrieved from database and if succeeded assign itself parsed values.
func (p *PostgresPoint) Scan(src interface{}) error {
  val, ok := src.([]byte)
  if !ok {
    return fmt.Errorf("value contains unexpected type")
  }
  _, err := fmt.Sscanf(string(val), "(%f,%f)", &p[0], &p[1])

  return err
}

const (
  // PointPrecision is a geo point precision.
  PointPrecision = 8
)

// Trunc truncates longitude and latitude to fixed precision..
func Trunc(point Point) Point {
  point[0] = util.Trunc(point[0], PointPrecision)
  point[1] = util.Trunc(point[1], PointPrecision)

  return point
}
