CREATE TABLE "users" (
  "national_id" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "permission" varchar[] NOT NULL,
  "has_voted" boolean NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "candidates" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "dob" varchar NOT NULL,
  "bio_link" varchar NOT NULL,
  "image_url" varchar NOT NULL,
  "policy" text NOT NULL,
  "vote_count" integer NOT NULL,
  "percentage" integer NOT NULL,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "votes" (
  "id" bigserial PRIMARY KEY,
  "vote_national_id" varchar NOT NULL,
  "candidate_id" bigserial NOT NULL,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "election_properties" (
  "id" bigserial PRIMARY KEY,
  "name" varchar UNIQUE NOT NULL,
  "value" boolean NOT NULL,
  "create_at" timestamptz NOT NULL DEFAULT (now())
);

INSERT INTO "election_properties" ("name", "value") VALUES ('ELECTION_CLOSED', 'f');

ALTER TABLE "votes" ADD FOREIGN KEY ("vote_national_id") REFERENCES "users" ("national_id");

ALTER TABLE "votes" ADD FOREIGN KEY ("candidate_id") REFERENCES "candidates" ("id");

ALTER TABLE "votes" ADD CONSTRAINT "vote_national_id_key" UNIQUE ("vote_national_id");


CREATE OR REPLACE FUNCTION vote_event_trigger_fnc()
  RETURNS trigger AS
$$
BEGIN
	UPDATE candidates SET vote_count = vote_count + 1, percentage = (
    	(
    		(select COUNT(*) from votes where candidate_id = NEW."candidate_id")/
    		(select COUNT(*) from users where 'VOTE'=ANY("permission"))::float
    	)*100
  )
	WHERE id = NEW."candidate_id";
  UPDATE users SET has_voted = 't'
  WHERE national_id = NEW."vote_national_id";
RETURN NEW;
END;
$$
LANGUAGE 'plpgsql';

CREATE TRIGGER vote_event_trigger
  AFTER INSERT
  ON "votes"
  FOR EACH ROW
  EXECUTE PROCEDURE vote_event_trigger_fnc();
