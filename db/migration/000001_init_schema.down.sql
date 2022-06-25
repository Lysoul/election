DROP TRIGGER vote_event_trigger on "votes" ;

DROP FUNCTION vote_event_trigger_fnc;

ALTER TABLE IF EXISTS "votes" DROP CONSTRAINT IF EXISTS "votes_vote_national_id_fkey";

DROP TABLE IF EXISTS election_properties;

DROP TABLE IF EXISTS votes;

DROP TABLE IF EXISTS candidates;

DROP TABLE IF EXISTS users;

