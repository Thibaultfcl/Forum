# FORUM 💻

This project is a web forum developed in Golang which allow the user to discuss topics that they like.


## Main features

- Authentication :
    - You can create an account to create posts and interact with other users
    - The connexion is saved in your cookies to stay connected
    - We use Universally Unique IDentifier to ensure the connection is secure

- Content :
    - Anyone can the the content
    - If you want to create posts or comments you need to create an account

- Moderation :
    - Only admin users can delete post, manage ban and moderator status or reset the profile picture
    - We have moderator that can report post to the admin

- Profile :
    - Once you created an account, you can modifie your information as you want
    - You can even change your profile picture

- Search :
    - You can type on the search bar to go to a specific page
    - There is suggestion based on the database

- Security :
    - The password is encrypt in the database
    - We have secure sessions with unique identifier

## HOW TO ACCESS IT

- Run this command in your terminal : `go run .\main.go`
- Then you can access the forum troughth this link [localhost:8080/home](http://localhost:8080/home)