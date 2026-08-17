import EditAndMoveTask from "../components/EditAndMoveTask";

export default function Home() {
  return (
    <main className="page-shell">
      <section className="composition-slot" aria-label="Task board application">
        <EditAndMoveTask />
      </section>
    </main>
  );
}
