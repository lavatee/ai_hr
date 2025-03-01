CREATE TABLE IF NOT EXISTS interviews
(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    stack VARCHAR(255) NOT NULL,
    difficulty VARCHAR(20) NOT NULL,
    current_task_index INT
);

CREATE TABLE IF NOT EXISTS tasks 
(
    id SERIAL PRIMARY KEY,
    interview_id INT NOT NULL,
    index INT, 
    correct_answer_index INT,
    text VARCHAR(511) NOT NULL
);

CREATE TABLE IF NOT EXISTS answers
(
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL,
    text VARCHAR(255),
    index INT
);

ALTER TABLE tasks ADD FOREIGN KEY (interview_id) REFERENCES interviews(id);
ALTER TABLE answers ADD FOREIGN KEY (task_id) REFERENCES tasks(id);