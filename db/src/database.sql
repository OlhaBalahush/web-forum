-- ALTER TABLE posts ADD COLUMN comments    JSONB;
-- DROP TABLE users;
-- DROP TABLE posts;
-- DROP TABLE comments;
-- DROP TABLE relations_likes;
-- DROP TABLE relations_likes_comments;
-- DROP TABLE relations_categories;
-- DROP TABLE cookies;
-- DROP TABLE notifications_requests;

SELECT * FROM posts;
SELECT * FROM comments;
SELECT * FROM relations_likes;
SELECT * FROM relations_likes_comments;
SELECT * FROM relations_categories;
SELECT * FROM cookies;
SELECT * FROM users;
SELECT * FROM notifications_requests;

DELETE FROM users WHERE user_name = "Ольга Балагуш";

SELECT id FROM users WHERE user_name = "Ольга Балагуш" AND role = 3;
-- SELECT id FROM notifications WHERE action = "request" AND who_did_it = ?

CREATE TABLE IF NOT EXISTS users (
        id 		INTEGER PRIMARY KEY AUTOINCREMENT,
        user_name 	TEXT NOT NULL,
        email TEXT 	NOT NULL UNIQUE,
        password 	TEXT NOT NULL,
	about 		TEXT NOT NULL,		
        image_path      TEXT NOT NULL,
        role            INTEGER NOT NULL,
        created_at 	DATE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS posts (
	id 		INTEGER PRIMARY KEY AUTOINCREMENT,
	title 		TEXT NOT NULL,
	text 		TEXT NOT NULL,
	user_id 	INTEGER REFERENCES users (id),   
	created_at 	DATE DEFAULT CURRENT_TIMESTAMP NOT NULL,
        image_path      TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS comments (
        id 		INTEGER PRIMARY KEY AUTOINCREMENT,
        text 	        TEXT NOT NULL,
        user_id 	INTEGER REFERENCES users (id),
        post_id 	INTEGER REFERENCES posts (id),
        created_at 	DATE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS relations_likes (
        id 	        INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id 	INTEGER REFERENCES users (id),
        post_id 	INTEGER REFERENCES posts (id),
        mark 	        INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS relations_likes_comments (
        id              INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id         INTEGER REFERENCES users (id),
        comment_id      INTEGER REFERENCES comments (id),
        mark            INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS relations_categories (
        post_id 	INTEGER REFERENCES posts (id),
        category 	TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cookies (
        user_id         INTEGER REFERENCES users (id),
        name            TEXT NOT NULL,
        value           TEXT NOT NULL UNIQUE,
        expires         DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS notifications_requests (
        id 		INTEGER PRIMARY KEY AUTOINCREMENT,
        reciver_id      INTEGER REFERENCES users (id),
        requestor_id    INTEGER REFERENCES users (id),
        action          TEXT NOT NULL,
        post_id         INTEGER REFERENCES posts (id),
        message         TEXT,
        seen            INTEGER DEFAULT 0,
        created_at 	DATE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

INSERT INTO users (user_name, email, password, about, image_path, role, created_at) VALUES
("John Doe", "johndoe@example.com", "1111", "Lorem ipsum dolor sit amet, consectetur adipiscing elit.", "", 1, "2011-03-15"),
("Jane Smith", "janesmith@example.com", "1111", "Pellentesque accumsan tincidunt lacus.", "", 1, "2012-03-14"),
("Bob Johnson", "bobjohnson@example.com", "1111", "Nulla facilisi. Donec nec lacus vitae nunc dictum malesuada.", "", 1, "2013-03-13"),
("Alex Brown", "alexbrown@example.com", "1111", "Sed non risus. Suspendisse lectus tortor, dignissim sit amet, adipiscing nec, ultricies sed, dolor.", "", 1, "2014-03-12"),
("Eva Green", "evagreen@example.com", "1111", "Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero.", "", 1, "2015-03-11"),
("Sam Wilson", "samwilson@example.com", "1111", "In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo.", "", 1, "2016-03-10"),
("Lisa Johnson", "lisajohnson@example.com", "1111", "Nullam dictum felis eu pede mollis pretium.", "", 1, "2017-03-09"),
("Mark Davis", "markdavis@example.com", "1111", "Aliquam tincidunt mauris eu risus.", "", 1, "2018-03-08"),
("Sarah Lee", "sarahlee@example.com", "1111", "Vestibulum auctor dapibus neque.", "", 1, "2019-03-07"),
("James Baker", "jamesbaker@example.com", "1111", "Nunc dignissim risus id metus.", "", 1, "2020-03-06"),
("Olivia Taylor", "oliviataylor@example.com", "1111", "Cras ornare tristique elit.", "", 1, "2021-03-05"),
("George Parker", "georgeparker@example.com", "1111", "Vivamus vestibulum nulla nec ante.", "",  1, "2022-03-04"),
("Emily Smith", "emilysmith@example.com", "1111", "Praesent placerat risus quis eros.", "", 1, "2023-03-03");

INSERT INTO posts (title, text, user_id, image_path) VALUES
("My First Post", "Lorem ipsum dolor sit amet, consectetur adipiscing elit.", 1, ""),
("My Second Post", "Pellentesque accumsan tincidunt lacus.", 2, ""),
("My Third Post", "Nulla facilisi. Donec nec lacus vitae nunc dictum malesuada.", 3, ""),
("My Fourth Post", "Sed non risus. Suspendisse lectus tortor, dignissim sit amet, adipiscing nec, ultricies sed, dolor.", 4, ""),
("My Fifth Post", "Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero.", 5, ""),
("My Sixth Post", "In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo.", 6, ""),
("My Seventh Post", "Nullam dictum felis eu pede mollis pretium.", 7, ""),
("My Eighth Post", "Aliquam tincidunt mauris eu risus.", 8, ""),
("My Ninth Post", "Vestibulum auctor dapibus neque.", 9, ""),
("My Tenth Post", "Nunc dignissim risus id metus.", 10, ""),
("My Eleventh Post", "Cras ornare tristique elit.", 1, ""),
("My Twelfth Post", "Vivamus vestibulum nulla nec ante.", 2, ""),
("My Thirteenth Post", "Praesent placerat risus quis eros.", 3, "");

INSERT INTO comments (text, user_id, post_id, created_at) VALUES
("Great post!", 2, 1, "2023-03-15"),
("I really enjoyed reading this.", 3, 1, "2023-03-15"),
("Thanks for sharing!", 1, 2, "2023-03-14"),
("I had a similar experience.", 2, 2, "2023-03-14"),
("Interesting thoughts!", 3, 3, "2023-03-13"),
("Great post, thanks for sharing!", 2, 1, "2023-03-12"),
("I totally agree with you!", 3, 1, "2023-03-11"),
("This is really insightful, I'll have to try it out!", 4, 2, "2023-03-10"),
("Thanks for the helpful tips!", 5, 2, "2023-03-09"),
("I had never thought of it that way before, thanks for sharing your perspective.", 6, 3, "2023-03-08"),
("Great article, really enjoyed reading it!", 7, 4, "2023-03-07"),
("Thanks for the informative post!", 8, 4, "2023-03-06"),
("This is really helpful, thank you!", 9, 5, "2023-03-05"),
("I'm definitely going to try this out, thanks for sharing!", 10, 6, "2023-03-04"),
("This is really interesting, thanks for sharing your research!", 1, 6, "2023-03-03");

INSERT INTO relations_likes (user_id, post_id, mark) VALUES
(1, 1, 1),
(2, 1, 1),
(3, 1, 1),
(1, 2, 1),
(2, 2, 1),
(1, 3, 1),
(1, 3, 1),
(2, 3, -1),
(3, 1, 1),
(4, 2, -1),
(5, 2, 1),
(6, 1, 1),
(7, 4, -1),
(8, 4, 1),
(9, 5, -1),
(10, 6, 1);

INSERT INTO relations_likes_comments (user_id, comment_id, mark) VALUES
(1, 1, 1),
(2, 1, 1),
(3, 2, 1),
(1, 3, 1),
(2, 4, 1),
(3, 5, 1),
(1, 1, 1),
(2, 2, -1),
(3, 3, 1),
(4, 4, -1),
(5, 5, 1),
(6, 6, 1),
(7, 7, -1),
(8, 8, 1),
(9, 9, -1),
(10, 10, 1);

INSERT INTO relations_categories (post_id, category) VALUES
(1, "Technology"),
(1, "Software"),
(2, "Food"),
(2, "Cooking"),
(3, "Sports"),
(4, "Travel"),
(4, "Adventure"),
(5, "Fashion"),
(6, "Art"),
(6, "Design"),
(7, "History"),
(7, "Politics"),
(8, "Education"),
(9, "Music"),
(9, "Entertainment"),
(10, "Fitness"),
(10, "Health"),
(11, "Business"),
(12, "Photography"),
(12, "Nature"),
(13, "Gaming"),
(13, "Technology"),
(13, "Innovation");