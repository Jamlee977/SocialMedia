const userPosts = document.getElementById("user_posts");
const followButton = document.getElementById("follow_button");
const unfollowButton = document.getElementById("unfollow_button");
const editButton = document.getElementById("edit_button");

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
    postAuthor.innerHTML = `<a href="/profiles/${post.authorId}" style="text-decoration: none; color: black">${post.author}</a>`;

    const hr = document.createElement("hr");

    postCard.appendChild(postAuthor);
    postCard.appendChild(postContent);
    postBody.appendChild(postCard);
    postBody.appendChild(hr);
    userPosts.appendChild(postBody);
}

followButton?.addEventListener("click", async () => {
    if (followButton.innerText === "Follow") {
        await followUser(followButton);
    } else if (followButton.innerText === "Unfollow") {
        await removeFollow(followButton);
    } else {
        console.error("Error: Follow button not found");
        window.location.reload();
    }
});

unfollowButton?.addEventListener("click", async () => {
    if (unfollowButton.innerText === "Follow") {
        await followUser(unfollowButton);
    } else if (unfollowButton.innerText === "Unfollow") {
        await removeFollow(unfollowButton);
    } else {
        console.error("Error: Unfollow button not found");
        window.location.reload();
    }
});

editButton?.addEventListener("click", async () => {
    const response = await fetch(`/settings/edit-profile`, {
        method: "GET",
    });

    if (response.ok) {
        // document.location.replace("/settings/edit-profile");
        window.location.href = "/settings/edit-profile";
    } else {
        alert(response.statusText);
    }
})

async function followUser(elementButton) {
    elementButton.innerText = "Unfollow";
    elementButton.classList.remove("btn-primary");
    elementButton.classList.add("btn-danger");
    elementButton.id = "unfollow_button";

    elementButton.disabled = true;

    const url = window.location.href;
    const userId = url.substring(url.lastIndexOf("/") + 1);
    await fetch(`/api/users/${userId}/follow`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
    });

    elementButton.disabled = false;
}

async function removeFollow(elementButton) {
    elementButton.innerText = "Follow";
    elementButton.classList.remove("btn-danger");
    elementButton.classList.add("btn-primary");
    elementButton.id = "follow_button";

    elementButton.disabled = true;

    const url = window.location.href;
    const userId = url.substring(url.lastIndexOf("/") + 1);
    await fetch(`/api/users/${userId}/unfollow`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
    });

    elementButton.disabled = false;
}
