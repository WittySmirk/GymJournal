// TODO: Create the modals



function ExerciseButton(props: { create: boolean, name?: string, id?: string }) {
  async function postToServer() {
    const d = {
      id: props.id,
      message: "hello buddy"
    };
    const resp = await fetch("http://localhost:8080/app", {
      method: "POST",
      body: JSON.stringify(d)
    })
    console.log(resp)
  }
  return <button onClick={postToServer} className="exerciseButton">
    {props.create ? "Create New" : props.name}
  </button>;
}

export default ExerciseButton;
