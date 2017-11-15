## ItsYou.online login user handling

What happens when a user logs in with ItsYou.online for the first time:

1. The user can link an existing gitea user with the same itsyou.online user.
2. or the user can create a new account on gitea to be linked with itsyou.online.

note that the user will be able to login through itsyou.online or through his normal username and password

## ItsYou.online organizations integration

It is possible to add ItsYou.online organizations as collaborators to repositories.
All users who have access to the organizations that have been added will then be
granted the corresponding access rights. These rights are evaluated every time the user
tries to perform an action such as pushing a commit. Because this evaluation requires
an API access key on ItsYou.online, only the organization previously defined in the
settings, and all of its children can be successfully authorized in this way. Therefore,
despite the ability to add any organization, only the aforementioned ones will be able
to authenticate successfully as members of the collaborating ItsYou.online organization.


## To extend locales of the application
Extend your locale under options/locale ( for example you will have `options/locale/locale_en-US.ini`)
Add required words

```
sign_in_itsyouonline = Sign in using ItsyouOnline

```
Then in templates you can access it using
```
{{.i18n.Tr "sign_in_itsyouonline" }}
```

## to extend templates of the application
You can create new directory in application custom dir `custom/templates` then create template files with the same name
e.g.: [Custom Gitea Template](`https://github.com/Incubaid/gitea_templates`)
