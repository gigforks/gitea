## Gitea Installation:
You can Install gitea normally from [official docs](https://docs.gitea.io/en-us/install-from-source/)

## ItsYou.online Integration

To Integrate itsyou.online with gitea you need:

* Use our gitea fork branch [iyo_integration](https://github.com/gigforks/gitea/tree/iyo_integration)

```bash
cd $GOPATH/src/code.gitea.io/gitea
git remote add gigfork https://github.com/gigforks/gitea.git
git pull gigfork 
git checkout iyo_cleanup
```

* Rebuild gitea
 ```bash
cd $GOPATH/src/code.gitea.io/gitea && TAGS="bindata" make generate build
```

* Add new oauth2 login source from database:
 ```
 insert into login_source (type, name, is_actived, cfg, created_unix, updated_unix) VALUES
  (6, 'Itsyou.onlineee', TRUE, 
  '{"Provider":"itsyou.online","ClientID":"abdoorg","ClientSecret":"f8G3fX9_4fIhfjxUjZrvFTPrObowRuL1i4iWnlS65HHd36m_9chD","OpenIDConnectAutoDiscoveryURL":"","CustomURLMapping":null}',
   extract('epoch' from CURRENT_TIMESTAMP) , extract('epoch' from CURRENT_TIMESTAMP)
  );
 ```

* Start Gitea
```bash
./gitea web
```

Note that when you will first login with itsyou.online, a user will be created on gitea linked to itsyou.online user. You will be able to choose the show username on gitea and set the password

## ItsYou.online organizations integration

It is possible to add ItsYou.online organizations as collaborators to repositories.
All users who have access to the organizations that have been added will then be
granted the corresponding access rights. These rights are evaluated every time the user
tries to perform an action such as pushing a commit. Because this evaluation requires
an API access key on ItsYou.online, only the organization previously defined in the
login source, and all of its children can be successfully authorized in this way. Therefore,
you can only add this organization or one of its children as collaborators, and only the aforementioned ones will be able
to authenticate successfully as members of the collaborating ItsYou.online organization.


## GIG cutom templates and locale
You can use GIG custom templates and locales by adding the contents of  [Custom Gitea](`https://github.com/Incubaid/gitea-custom`) repo into application custom dir `custom/`

## Migration from GOGS
You can use migration steps from official [gitea docs](https://docs.gitea.io/en-us/upgrade-from-gogs/)

##### * for integrating iyo users with the migrated gogs you will need to:
 * Add new oauth2 login source from database (The same way added in the new installation steps)
 * You will need to edit `user` table to link users with iyo:
  ```
   UPDATE "user"
   set login_name = lower_name
   where type = 0;
   ```
