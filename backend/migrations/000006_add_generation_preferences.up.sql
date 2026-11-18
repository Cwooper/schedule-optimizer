-- JSON of the request's validated preferences; NULL when the request omitted
-- them and the defaults applied.
ALTER TABLE generation_logs ADD COLUMN preferences TEXT;
