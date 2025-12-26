ALTER TABLE "picuinha_entries"
  ADD COLUMN "payment_method_id" int,
  ADD COLUMN "card_owner" varchar(20) NOT NULL DEFAULT 'SELF';

ALTER TABLE "picuinha_entries"
  ADD FOREIGN KEY ("payment_method_id") REFERENCES "payment_methods" ("payment_method_id");
