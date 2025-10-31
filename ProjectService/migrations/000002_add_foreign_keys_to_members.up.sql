ALTER TABLE project.projects
    ADD CONSTRAINT fk_projects_manager
    FOREIGN KEY (manager_id) 
    REFERENCES project.users_meta(user_id) 
    ON DELETE SET NULL;

ALTER TABLE project.tasks
    ADD CONSTRAINT fk_tasks_creator
    FOREIGN KEY (creator_id) 
    REFERENCES project.users_meta(user_id) 
    ON DELETE RESTRICT;

ALTER TABLE project.tasks
    ADD CONSTRAINT fk_tasks_assignee
    FOREIGN KEY (assign_id) 
    REFERENCES project.users_meta(user_id) 
    ON DELETE SET NULL;

ALTER TABLE project.project_members
    ADD CONSTRAINT fk_project_members_user
    FOREIGN KEY (user_id) 
    REFERENCES project.users_meta(user_id) 
    ON DELETE CASCADE;
