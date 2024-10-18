+++ 
date = 2024-03-12
title = "Django - migrating to a custom user model mid-project"
author = "Antti Viitala"
tags = [
    "python",
    "django"
]
+++

This was quite painful. Please just follow the best practice of [using a custom user model from the start.](https://docs.djangoproject.com/en/dev/topics/auth/customizing/#substituting-a-custom-user-model)

## Starting point in my app

- Default user model
- One model proxying the current user model, because I didn't want to do this change before
- Wanted to add a new app to contain the new custom user model (`shared`)
- Using `django-tenants`

## Process

### First stage

1. Create a new app, `shared`, to contain the new user model `User`. Just with `django-admin startapp shared`. Clean up files that won't be used like `views.py` etc.
2. Update `settings.py`: Add this app to **all tenants**
3. Before making any changes to models: Create a **single, blank migration** for this app, i.e. `0001_initial.py` that does nothing (`python manage.py makemigrations --empty shared`)
4. Run `python manage.py migrate` to get this first initial app set up.

**Here, between the stages, you should push these changes and run the migrations against your production and any other DBs that you wish to persists.**

### Second stage

1. Now update the `models.py` in `shared` app with your new model, like the one below. **Do not make any other changes at this point preferably**, if you want to add new fields, do it later once the change process is complete.

    ```python
    from django.db import models
    from django.contrib.auth.models import AbstractUser

    class User(AbstractUser):
        class Meta:
            db_table = 'auth_user'
    ```

1. Edit the `shared/migrations/0001_initial.py` migration that you had created before to handle the creation of the entire `User` model. Quoting below from one of the references on what to do:
    1. **This is because you want the initial migration to already have been executed in an instance with data (hence the first stage), but in the future you'll want to recreate it completely in new environments (e.g. in testing)**
    1. (The exact code to use here will likely change over time with newer versions of Django. You can find the current code by creating a new app temporarily, add the User model to it and then look at the migration file `./manage.py makemigrations` produces.)
1. Update the Django setting to use your new model `AUTH_USER_MODEL = "shared.User"`
1. Create a new migration with `manage.py makemigrations --empty shared`
1. Edit that second migration, called something like `shared/migrations/0002_xx.py`, adding a function that will convert the user model entries from the old model to the new one:

    ```python
    def change_user_type(apps, schema_editor):
        ContentType = apps.get_model('contenttypes', 'ContentType')
        ct = ContentType.objects.filter(
            app_label='auth',
            model='user'
        ).first()
        if ct:
            ct.app_label = 'user'
            ct.save()
    ```

1. **Now run your second migration with `manage.py migrate`**
1. All done! Double check everything, including logging in and the various pages of your app.

## Additional issues I encountered

- Wrong default manager in the new `User` model didn't allow me to log in, due to previously screwing up with the managers of my proxy model. (default query was returning active, non-staff users only, hence my superuser logins were not working)
- Postgres versions in different envs. `dbbackup` needs a matching Postgres version in order to function correctly.
- Previous references to `ActiveTenantUser` via a relationship from `VotingEvent` needed to be manually changed. The table column name of `activetenantuser_id` wouldn't change, so needed to do this manually in postgres in Prod

## References

- [Primary reference](https://code.djangoproject.com/ticket/25313#comment:24)
- [Additional notes for when you are starting with a blank new app as I did](https://code.djangoproject.com/ticket/25313#comment:27)
