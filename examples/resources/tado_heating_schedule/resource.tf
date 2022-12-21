# The following example shows how to create a heating schedule with a
# Monday - Sunday timetable.

resource "tado_heating_schedule" "kitchen" {
  home_name = "My Home"
  zone_name = "Kitchen"

  mon_sun = [
    { heating = false, start = "00:00", end = "06:00" },
    { heating = true, temperature = 20.0, start = "06:00", end = "21:00" },
    { heating = false, start = "21:00", end = "00:00" },
  ]
}

# The following example shows how to create a heating schedule with a
# Monday - Friday timetable and separate timetables for Saturday and Sunday.

resource "tado_heating_schedule" "living_room" {
  home_name = "My Home"
  zone_name = "Living Room"

  mon_fri = [
    { heating = false, start = "00:00", end = "06:00" },
    { heating = true, temperature = 18.0, start = "06:00", end = "17:00" },
    { heating = true, temperature = 20.0, start = "17:00", end = "22:00" },
    { heating = false, start = "22:00", end = "00:00" },
  ]

  sat = [
    { heating = false, start = "00:00", end = "08:00" },
    { heating = true, temperature = 20.0, start = "08:00", end = "23:00" },
    { heating = false, start = "23:00", end = "00:00" },
  ]

  sun = [
    { heating = false, start = "00:00", end = "08:00" },
    { heating = true, temperature = 20.0, start = "08:00", end = "22:00" },
    { heating = false, start = "22:00", end = "00:00" },
  ]
}

# The following example shows how to create a heating schedule with different
# timetables for each weekday.

resource "tado_heating_schedule" "bedroom" {
  home_name = "My Home"
  zone_name = "Bedroom"

  mon = [
    { heating = false, start = "00:00", end = "06:00" },
    { heating = true, temperature = 20.0, start = "06:00", end = "09:00" },
    { heating = true, temperature = 18.0, start = "09:00", end = "20:00" },
    { heating = true, temperature = 20.0, start = "20:00", end = "22:00" },
    { heating = false, start = "22:00", end = "00:00" },
  ]

  tue = [
    { heating = false, start = "00:00", end = "06:00" },
    { heating = true, temperature = 20.0, start = "06:00", end = "09:00" },
    { heating = true, temperature = 18.0, start = "09:00", end = "20:00" },
    { heating = true, temperature = 20.0, start = "20:00", end = "22:00" },
    { heating = false, start = "22:00", end = "00:00" },
  ]

  wed = [
    { heating = false, start = "00:00", end = "06:00" },
    { heating = true, temperature = 20.0, start = "06:00", end = "09:00" },
    { heating = true, temperature = 18.0, start = "09:00", end = "20:00" },
    { heating = true, temperature = 20.0, start = "20:00", end = "22:00" },
    { heating = false, start = "22:00", end = "00:00" },
  ]

  thu = [
    { heating = false, start = "00:00", end = "06:00" },
    { heating = true, temperature = 20.0, start = "06:00", end = "09:00" },
    { heating = true, temperature = 18.0, start = "09:00", end = "20:00" },
    { heating = true, temperature = 20.0, start = "20:00", end = "22:00" },
    { heating = false, start = "22:00", end = "00:00" },
  ]

  fri = [
    { heating = false, start = "00:00", end = "06:00" },
    { heating = true, temperature = 20.0, start = "06:00", end = "09:00" },
    { heating = true, temperature = 18.0, start = "09:00", end = "20:00" },
    { heating = true, temperature = 20.0, start = "20:00", end = "00:00" },
  ]

  sat = [
    { heating = false, start = "00:00", end = "08:00" },
    { heating = true, temperature = 20.0, start = "08:00", end = "11:00" },
    { heating = true, temperature = 18.0, start = "11:00", end = "20:00" },
    { heating = true, temperature = 20.0, start = "20:00", end = "00:00" },
  ]

  sun = [
    { heating = false, start = "00:00", end = "08:00" },
    { heating = true, temperature = 20.0, start = "08:00", end = "11:00" },
    { heating = true, temperature = 18.0, start = "11:00", end = "20:00" },
    { heating = true, temperature = 20.0, start = "20:00", end = "22:00" },
    { heating = false, start = "22:00", end = "00:00" },
  ]
}


