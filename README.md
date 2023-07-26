# forum-moderation

## Table of Contents
- ### [General Information](#general-information)
- ### [Technologies](#technologies)
- ### [Setup](#setup)
- ### [Usage](#usage)
- ### [Authors](#authors)

## General Information
 The project involves creating a web forum where users can communicate through posts and comments, with the ability to like and dislike posts and filter content by categories, created posts, and liked posts. User authentication is required for creating and interacting with content, with passwords being encrypted and stored in an SQLite database. The use of cookies is necessary for user login sessions, and Docker is required for containerization. The project also involves learning about web basics, SQL language, encryption, and using and setting up Docker.

 In this project we have at least 4 types of users :

- Guests
    These are unregistered-users that can neither post, comment, like or dislike a post. They only have the permission to see those posts, comments, likes or dislikes.
- Users
    These are the users that will be able to create, comment, like or dislike posts.
- Moderators
    Moderators, as explained above, are users that have a granted access to special functions :
    - They should be able to monitor the content in the forum by deleting or reporting post to the admin
    To create a moderator the user should request an admin for that role
- Administrators
    Users that manage the technical details required for running the forum. This user must be able to :
    - Promote or demote a normal user to, or from a moderator user.
    - Receive reports from moderators. If the admin receives a report from a moderator, he can respond to that report
    - Delete posts and comments
    - Manage the categories, by being able to create and delete them.

## Technologies
- ### [Golang](https://go.dev/)
- ### [HTML](https://www.w3.org/html/)
- ### [CSS](https://developer.mozilla.org/en-US/docs/Web/CSS)
- ### [Javascript](https://www.javascript.com/)
- ### [SQLite](https://sqlite.org/index.html)

## Setup
Clone the repository
```
git clone https://01.kood.tech/git/Olya/forum-moderation.git
```

## Usage
### Without docker
Run <code>.go</code> file from <code>../forum</code> folder
```
go run . 
```

For audit:

Administrator:

```
admin@gmail.com
```
```
1111
```

Moderator:

```
mod@gmail.com
```
```
1111
```

### With docker
Run <code>runDocker.sh</code> file from <code>../forum</code> folder
```
bash runDocker.sh
```
When you finish the audit, run <code>stopDocker.sh</code> file from <code>../forum</code> folder
```
bash stopDocker.sh
```

You can test the project from the page [https://localhost:8080/](https://localhost:8080/)


## Authors
- [Olha Balagush](https://01.kood.tech/git/Olya)
- [Elena Khaletska](https://01.kood.tech/git/ekhalets)
- [Taivo Tokman](https://01.kood.tech/git/TaivoT)
- [Glib Ovcharenko](https://01.kood.tech/git/govchare)