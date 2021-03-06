CREATE EXTENSION IF NOT EXISTS cube;
CREATE EXTENSION IF NOT EXISTS earthdistance;

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now();
RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE FUNCTION fix_point_precision()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.point = point(TRUNC(NEW.point[0]::numeric, 8), TRUNC(NEW.point[1]::numeric, 8));
RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE users (
    id SERIAL,
    username varchar(16),
    created_at timestamp DEFAULT current_timestamp NOT NULL,
    updated_at timestamp DEFAULT current_timestamp NOT NULL,

    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_username_key UNIQUE (username),
    CONSTRAINT users_username_valid CHECK (username ~ '^[a-zA-Z0-9]{4,16}$')
);

CREATE TABLE locations (
    user_id INT,
    point POINT,
    created_at timestamp DEFAULT current_timestamp NOT NULL,
    updated_at timestamp DEFAULT current_timestamp NOT NULL,

    CONSTRAINT locations_pkey PRIMARY KEY (user_id),
    CONSTRAINT locations_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT locations_longitude_valid CHECK (point[0] >= -180 AND point[0] <= 180),
    CONSTRAINT locations_latitude_valid CHECK (point[1] >= -90 AND point[1] <= 90)
);

CREATE TRIGGER update_updated_at BEFORE UPDATE
    ON users FOR EACH ROW EXECUTE PROCEDURE
    update_updated_at();

CREATE TRIGGER update_updated_at BEFORE UPDATE
    ON locations FOR EACH ROW EXECUTE PROCEDURE
        update_updated_at();

CREATE TRIGGER fix_points_precision BEFORE INSERT OR UPDATE
    ON locations FOR EACH ROW EXECUTE PROCEDURE
        fix_point_precision();
