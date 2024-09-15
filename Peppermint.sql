CREATE TABLE users (
  "user_id" SERIAL UNIQUE NOT NULL,
  "username" VARCHAR(64) UNIQUE NOT NULL,
  "pw_hash" VARCHAR(1024) NOT NULL,
  PRIMARY KEY ("user_id", "username")
);

CREATE TABLE transactions (
  "teller_id" VARCHAR(64) UNIQUE PRIMARY KEY NOT NULL,
  "merchant" TEXT NOT NULL,
  "description" TEXT NOT NULL,
  "posted_date" DATE NOT NULL,
  "amount" MONEY NOT NULL,
  "category" BIGSERIAL NOT NULL,
  "user_id" SERIAL NOT NULL
);

CREATE TABLE saved_categories (
  "user_id" SERIAL NOT NULL,
  "category_id" BIGSERIAL PRIMARY KEY,
  "category_name" VARCHAR(128) NOT NULL,
  "category_limit" MONEY NOT NULL
);

CREATE TABLE user_finances (
  "user_id" SERIAL NOT NULL,
  "amt_401k_contribution" MONEY NOT NULL DEFAULT 0,
  "total_insurance_amount" MONEY NOT NULL DEFAULT 0,
  "monthly_posttax_salary" MONEY NOT NULL
);

CREATE TABLE recurring_costs (
  "user_id" SERIAL NOT NULL,
  "name" VARCHAR(128) NOT NULL,
  "amount" MONEY NOT NULL,
  "month_frequency" INTEGER NOT NULL,
  "is_savings" BOOL NOT NULL
);

CREATE TABLE onetime_costs (
  "user_id" SERIAL NOT NULL,
  "name" VARCHAR(128) NOT NULL,
  "amount" MONEY NOT NULL,
  "month" SMALLINT NOT NULL,
  "year" SMALLINT NOT NULL
);

ALTER TABLE transactions ADD FOREIGN KEY ("category") REFERENCES "saved_categories" ("category_id");

ALTER TABLE transactions ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id");

ALTER TABLE saved_categories ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id");

ALTER TABLE user_finances ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id");

ALTER TABLE recurring_costs ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id");

ALTER TABLE onetime_costs ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id");
