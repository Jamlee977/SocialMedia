const userPosts = document.getElementById("user_posts");
const followButton = document.getElementById("follow_button");
const unfollowButton = document.getElementById("unfollow_button");

window.onload = async () => {
    const url = window.location.href;
    const userId = url.substring(url.lastIndexOf("/") + 1);
    const userPosts = await fetch(`/api/posts/${userId}`, {
        method: "GET",
        headers: {
            "Content-Type": "application/json",
        },
    });

    const userPostsData = await userPosts.json();

    if (!userPosts.ok) {
        return;
    }

    userPostsData.forEach((post) => {
        createPostElement(post);
    });
}

function createPostElement(post) {
    const postBody = document.createElement("div");
    postBody.classList.add("user_post_body");

    const postCard = document.createElement("div");
    postCard.classList.add("card-body");

    const postContent = document.createElement("p");
    postContent.classList.add("card-text");
    postContent.classList.add("text-justify");
    postContent.innerText = post.content;

    const postAuthor = document.createElement("h5");
    postAuthor.classList.add("card-text");
    postAuthor.classList.add("text-right");
    postAuthor.innerText = post.author;

    const hr = document.createElement("hr");

    postCard.appendChild(postAuthor);
    postCard.appendChild(postContent);
    postBody.appendChild(postCard);
    postBody.appendChild(hr);
    userPosts.appendChild(postBody);
}

followButton?.addEventListener("click", async () => {
    followButton.innerText = "Unfollow";
    followButton.classList.remove("btn-primary");
    followButton.classList.add("btn-danger");
    followButton.id = "unfollow_button";

    const url = window.location.href;
    const userId = url.substring(url.lastIndexOf("/") + 1);
    await fetch(`/api/users/${userId}/follow`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
    });
});

unfollowButton?.addEventListener("click", async () => {
    unfollowButton.innerText = "Follow";
    unfollowButton.classList.remove("btn-danger");
    unfollowButton.classList.add("btn-primary");
    unfollowButton.id = "follow_button";

    // const url = window.location.href;
    // const userId = url.substring(url.lastIndexOf("/") + 1);
    // await fetch(`/api/users/${userId}/unfollow`, {
    //     method: "POST",
    //     headers: {
    //         "Content-Type": "application/json",
    //     },
    // });
});
