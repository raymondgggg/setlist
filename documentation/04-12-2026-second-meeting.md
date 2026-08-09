# Meeting Objectives

- [x] Discuss what the work items we need to complete in order to get both projects to a state where we can start working on features.
- [x] Discuss roughly what the features are that we want.

# Meeting Objective 1

- Things our project needs to have:
  - Backend
    - ORM (GORM)
      - An example entity (Postgres table & DAO/model)
    - Ability to add an automated test (just configure it at this point)
    - Logger (uber's zap)
    - Migrations (tern)
    - Middleware
      - GraphQL
      - Session
      - Permission checking
    - Authentication
      - OAuth
        - Google
        - Github
        - Spotify
      - JWT
        - Frontend will pass this token when making POST requests to GQL
    - Add DB container
    - Service function(s) to connect to the Spotify API
    - Environment variables/files
  - Frontend
    - Tanstack npm package install
      - Configure UI project to use tanstack routing
      - Add a default home page using tanstack workflow (tanstack query, routing, etc.)
    - Add Component folder to project
    - Ability to add an automated test (just configure it at this point)
    - Logging

# Meeting Objective 2

- What features do we want?
  - Users
  - Friends list (follow people)
  - The user needs the ability to make a list of songs/albums that they’ve listened to.
    - The user needs the ability to make lists & add songs or albums to these lists. (users can add both songs & albums to a single list - tentative).
  - The user needs the ability to rank songs on an album
  - The user needs the ability to add a review to a particular song or album
    - Or songs within the album too?
  - The user needs the ability to comment on another user’s review.
  - A user’s profile settings should allow them to configure the theme they want, Light/Dark. Upon selection, the theme the user chose is used throughout the site.
  - The user needs the ability to be able to see their profile, as well as other user’s profiles.
  - “Watchlist” A list of things that you want to listen to (songs or albums)
  - Add “Song” or “Album” profile page. Essentially acts as the page you see when you click into a movie on letterboxd. In letterboxd, this page shows the movie poster, cast, description, and other details, this page also has buttons that allows the user to add a movie to their watchlist or review a movie, we’d want the mega equivalent for songs on our app. (does Genius have an api?)
  - Home or landing page when you first login, show all the popular albums, there’s a list of the most popular reviews for a given time period, popular list
