## gitea_testing
This document describes all the manual tests to be executed in order to validate gitea 

### Test 1 (test full text search in tickets cross repositories on which you have access)
#### Steps 
- test search on issue title
- search on issue comment
- search on users
- search on wiki contents
- search on repo / organization names
#### Results 
Results should be url's with a bit of context (like google):
- organizations
- repos
- users
- wiki pages
- issues

### Test 2 (test adding new collaberator to repository)
#### scenario 1( add organization which doesn't belong to the main organization)
- Go to one of your repository .
- try to add new itsYou.online organization which doesn't belong to the main organization,it should raise error
``` 
the organization you have added doesn't belong to the main organization 
```
#### Scenario 2 (add organization with read access)
- Go to one of your repository.
- Try to add with read access new itsYou.online organization which belong to main organization , and user1 have access to it . 
- Check that user1 can access this repository but can't edit or create issue in this repository .
- Check that user1 can clone git repo over ssh ,but can't push .
- Delete this organization from  the repository.
- Check that user1 can't access the repository anymore.
- Check that user1 can't clone git repo over ssh and can't push anymore.

#### Scenario 3 (add organization with write access)
- Go to one of your repository.
- Try to add with write access new itsYou.online organization which belong to main organization , and user1 have access to it.
- Check that user1 can access this repository and can add files and edit issues ,but can't use settings .
- Check that user1 can clone git repo over ssh and push edits .
- Check that you can assign issue from this repo to user1.
- Delete this organization from  the repository.
- Check that user1 can't access the repository anymore.
- Check that user1 can't clone git repo over ssh and can't push anymore.

#### Scenario 4 (add organization with admin access )
- Go to one of your repository.
- Try to add with admin access new itsYou.online organization which belong to main organization , and user1 have access to it.
- Check that user1 can access this repository and can add files ,edite issues and and use settings.


### Test 3 (Delete organization from itsyouonline side)
#### Scenario
- Go to one of your repository.
- Try to add with write access new itsYou.online organization which belong to main organization , and user1 have access to it.
- Delete this organization from user1 organizations from itsyouonline.
- logout the login again in gitea with user1.
- Check that user1 can't access the repository anymore.
- Check that user1 can't clone git repo over ssh and can't push anymore.

### Test4 (Test adding labels and milestones to issues)
#### Scenario
- Go to one of your repository.
- Try to add label and milestone to issue .

### Test5 (Test that you can login only through itsyouonline )
#### Scenario
- login always should redirect you to itsyouonline authentication page .
- logout and try to create or comment on issue in public repo , should redirect you to itsyouonline authentication page .
